package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
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
		viewport viewport.Model
		textarea textarea.Model
		messages []string
		con      net.Conn
		err      error
	}
)

type MessageFrom int

const (
	FromYou = iota
	FromOther
	FromRoom
)

var chatStyles = map[MessageFrom]lipgloss.Style{
	FromYou:   lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
	FromOther: lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
	FromRoom:  lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message:"
	ta.Focus()

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

	return model{
		textarea: ta,
		viewport: vp,
		messages: []string{},
		err:      nil,
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
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case msgReceived:
		m.AddMessage(msg.from, msg.value, chatStyles[msg.fromType])
		m.RenderMessages()
		m.viewport.GotoBottom()
	case registerCon:
		m.con = msg.con
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)
		if len(m.messages) > 0 {
			m.RenderMessages()
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			if m.textarea.Value() == "" {
				return m, nil
			}

			if m.con != nil {
				m.con.Write([]byte(m.textarea.Value() + "\n"))
			}
			m.AddMessage("You", m.textarea.Value(), chatStyles[FromYou])
			m.RenderMessages()
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf("%s%s%s", m.viewport.View(), gap, m.textarea.View())
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}
