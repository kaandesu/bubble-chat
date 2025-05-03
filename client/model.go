package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n"

type homepage struct {
	focusIndex int
	nameInput  textinput.Model
	registered bool
	name       string
}

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
		viewport     viewport.Model
		textarea     textarea.Model
		homepage     homepage
		screenWidth  int
		screenHeight int
		activePage   int
		messages     []string
		con          net.Conn
		err          error
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
)

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

	return model{
		textarea:     ta,
		viewport:     vp,
		messages:     []string{},
		err:          nil,
		activePage:   0,
		screenWidth:  50,
		screenHeight: 50,
		homepage: homepage{
			focusIndex: 0,
			nameInput:  ni,
			registered: false,
			name:       "",
		},
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
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		niCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	m.homepage.nameInput, niCmd = m.homepage.nameInput.Update(msg)

	switch msg := msg.(type) {
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
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)
		if len(m.messages) > 0 {
			m.RenderMessages()
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			if m.homepage.focusIndex == 1 {
				m.homepage.nameInput.Focus()
				m.homepage.focusIndex = 0
			} else {
				m.homepage.nameInput.Blur()
				m.homepage.focusIndex = 1
			}
		case tea.KeyDown:
			if m.activePage == 0 {
				m.activePage = 1
				m.textarea.Focus()
				m.homepage.nameInput.Blur()
			} else {
				m.activePage = 0
				m.textarea.Blur()
				m.homepage.nameInput.Focus()
			}
			return m, nil
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			if (m.activePage == 1 && m.textarea.Value() == "") || (m.activePage == 0 && m.homepage.nameInput.Value() == "") {
				return m, nil
			}

			if m.activePage == 0 {
				if m.homepage.nameInput.Value() != "" {
					if m.homepage.focusIndex == 1 {
						m.homepage.name = m.homepage.nameInput.Value()
						m.activePage = 1
						m.textarea.Focus()
						m.homepage.nameInput.Blur()
					} else {
						m.homepage.nameInput.Blur()
						m.homepage.focusIndex = 1
					}
				}
			} else if m.activePage == 1 {
				if m.con != nil {
					m.con.Write([]byte(m.homepage.name + ">>" + m.textarea.Value() + "\n"))
				}
				m.AddMessage(m.homepage.name+"(You)", m.textarea.Value(), chatStyles[FromYou])
				m.RenderMessages()
				m.textarea.Reset()
				m.viewport.GotoBottom()

			}

		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd, niCmd)
}

func (m *model) RenderHomepage() string {
	title := "Enter your username!"
	titleStyle := lipgloss.NewStyle().Bold(true).MarginBottom(1)
	buttonStyle := lipgloss.NewStyle().Bold(true).MarginTop(1)
	var button string
	if m.homepage.focusIndex == 0 {
		button = buttonStyle.Render(blurredButton)
	} else {
		button = buttonStyle.Render(focusedButton)
	}

	form := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(title),
		m.homepage.nameInput.View(),
		button,
	)

	centered := lipgloss.Place(
		m.screenWidth,
		m.screenHeight,
		lipgloss.Center,
		lipgloss.Center,
		form,
	)

	return centered
}

func center(s string, w int) string {
	if len(s) >= w {
		return s
	}
	n := w - len(s)
	div := n / 2
	return strings.Repeat(" ", div) + s + strings.Repeat(" ", div)
}

func (m model) View() string {
	switch m.activePage {
	case 0:
		return m.RenderHomepage()
	case 1:
		return fmt.Sprintf("%s%s%s", m.viewport.View(), gap, m.textarea.View())
	default:
		return fmt.Sprintf("You are not suppose to see this!")
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tea.SetWindowTitle("Bubble-Chat"))
}
