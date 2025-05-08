package main

import (
	messages "bubble-client/messages"
	"bubble-client/styles"
	"bubble-client/views"
	"fmt"
	"net"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n"

type (
	errMsg      error
	msgReceived struct {
		from     string
		value    string
		fromType MessageFrom
	}
	registerCon struct {
		con net.Conn
	}
	model struct {
		homepage     views.Model
		roomlist     list.Model
		viewport     viewport.Model
		username     string
		textarea     textarea.Model
		screenWidth  int
		screenHeight int
		connected    bool
		activePage   int
		messages     []string
		con          net.Conn
		err          error
		infoText     string
	}
)

type MessageFrom int

const (
	FromYou = iota
	FromOther
	FromRoom
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	focusedButton = focusedStyle.Render("[ Register ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Register"))

	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#e9d502"))
)

func warningText(text string) string {
	return fmt.Sprintf("<< %s >>", warningStyle.Render(text))
}

var chatStyles = map[MessageFrom]lipgloss.Style{
	FromYou:   lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
	FromOther: lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
	FromRoom:  lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message:"

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	/* Login View */
	ni := textinput.New()
	ni.Placeholder = "Name"
	ni.Focus()
	ni.CharLimit = 156
	ni.Width = 20

	list := initTestList()

	return model{
		homepage:     views.InitialModel(),
		roomlist:     list,
		textarea:     ta,
		viewport:     vp,
		messages:     []string{},
		err:          nil,
		infoText:     "",
		connected:    false,
		activePage:   0,
		screenWidth:  50,
		screenHeight: 50,
	}
}

func (m *model) AddMessage(sender, payload string, style lipgloss.Style) {
	m.messages = append(m.messages, style.Render(sender+": ")+payload)
}

func (m *model) RenderMessages() {
	// wrap the content before setting it
	m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd       tea.Cmd
		vpCmd       tea.Cmd
		niCmd       tea.Cmd
		roomListCmd tea.Cmd
		homepageCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	m.roomlist, roomListCmd = m.roomlist.Update(msg)

	updated, homepageCmd2 := m.homepage.Update(msg)
	homepageCmd = homepageCmd2
	m.homepage = updated.(views.Model)

	switch msg := msg.(type) {
	case messages.NavigateTo:
		m.activePage = msg.To

		if m.activePage == 0 {
			m.textarea.Focus()
			m.roomlist.FilterInput.Blur()
		} else if m.activePage == 1 {
			updated, _ = m.homepage.Update(messages.SetActive{Value: false})
			m.homepage = updated.(views.Model)
			m.textarea.Focus()
			m.roomlist.FilterInput.Blur()
		} else if m.activePage == 2 {
			m.textarea.Blur()
			m.roomlist.FilterInput.Focus()
		}
		return m, nil
	case connectionStatus:
		m.connected = msg.err == nil
		if !m.connected {
			m.infoText = warningText("You are not connected to the server! Click [ctrl + r] to reconnect!")
		}
		return m, nil
	case msgReceived:
		m.AddMessage(msg.from, msg.value, chatStyles[msg.fromType])
		m.RenderMessages()
		m.viewport.GotoBottom()
	case registerCon:
		m.con = msg.con
	case tea.WindowSizeMsg:
		m.screenWidth = msg.Width
		m.screenHeight = msg.Height - lipgloss.Height(gap)

		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap) - lipgloss.Height(m.infoText)
		if len(m.messages) > 0 {
			m.RenderMessages()
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlR:
			if m.connected {
				return m, nil
			}
		case tea.KeyShiftDown:
			m.activePage++
			if m.activePage == 3 {
				m.activePage = 0
			}
			if m.activePage == 0 {
				m.textarea.Focus()
				// m.homepage.nameInput.Blur()
				m.roomlist.FilterInput.Blur()
			} else if m.activePage == 1 {
				m.textarea.Blur()
				m.roomlist.FilterInput.Blur()
				// m.homepage.nameInput.Focus()
			} else if m.activePage == 2 {
				// m.homepage.nameInput.Blur()
				m.textarea.Blur()
				m.roomlist.FilterInput.Focus()
			}
			return m, nil
		case tea.KeyRunes:
			if msg.String() == "q" {
				return m, nil
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			if m.activePage == 2 {
				addItemToList(&m.roomlist, "TEST ROOM")
			}
			if m.activePage == 1 && m.textarea.Value() == "" {
				return m, nil
			}

			if m.activePage == 1 {
				if m.con != nil {
					// TODO: send on enter
					// m.con.Write([]byte(m.homepage.name + ">>" + m.textarea.Value() + "\n"))
				}
				// m.AddMessage(m.homepage.name+"(You)", m.textarea.Value(), chatStyles[FromYou])
				m.AddMessage("(You)", m.textarea.Value(), chatStyles[FromYou])
				m.RenderMessages()
				m.textarea.Reset()
				m.viewport.GotoBottom()

			}

		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd, niCmd, homepageCmd, roomListCmd)
}

func (m *model) RenderHomepage() string {
	return m.homepage.View()
}

func (m model) View() string {
	switch m.activePage {
	case 0:
		return lipgloss.JoinVertical(
			lipgloss.Center,
			styles.Gap,
			m.infoText,
			m.RenderHomepage(),
		)
	case 1:
		return fmt.Sprintf("%s%s%s%s%s", m.infoText, gap, m.viewport.View(), gap, m.textarea.View())
	case 2:
		return fmt.Sprintf("\n\n\n\n[%s]\n\n[conn: %+v]\n\nON this page you see the list of stuff\n\n%s", m.infoText, m.connected, m.roomlist.View())
	default:
		return fmt.Sprintf("You are not suppose to see this!")
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tea.SetWindowTitle("Bubble-Chat"))
}
