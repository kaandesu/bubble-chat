package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const serverAddr = "127.0.0.1:3000"

type (
	errMsg      error
	msgReceived struct {
		from  string
		value string
	}
	registerCon struct {
		con net.Conn
	}
	model struct {
		viewport      viewport.Model
		textarea      textarea.Model
		messages      []string
		senderStyle   lipgloss.Style
		recieverStyle lipgloss.Style
		roomStyle     lipgloss.Style
		con           net.Conn
		err           error
	}
)

const gap = "\n\n"

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
		textarea:      ta,
		viewport:      vp,
		messages:      []string{},
		senderStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		recieverStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
		roomStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		err:           nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
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
		switch msg.from {
		case "Room":
			m.messages = append(m.messages, m.roomStyle.Render(msg.from+": ")+msg.value)
		default:
			m.messages = append(m.messages, m.recieverStyle.Render(msg.from+": ")+msg.value)
		}
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
	case registerCon:
		m.con = msg.con
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)
		if len(m.messages) > 0 {
			// wrap the content before setting it
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
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
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.textarea.Value())
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
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

func Connect(addr string, p *tea.Program) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	temp := make(chan struct{})

	con, err := net.DialTCP("tcp", nil, tcpAddr)
	if err == nil {
		p.Send(registerCon{
			con: con,
		})
	}
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := con.Read(buf)
			if err != nil {
				break
			}
			if n < 1 {
				continue
			}
			splits := strings.Split(string(buf[:n-1]), ">>")
			from, payload := splits[0], splits[1]
			p.Send(msgReceived{from: from, value: payload})
		}
	}()
	<-temp
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	go Connect(serverAddr, p)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
