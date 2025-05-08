package views

import (
	"bubble-client/messages"
	"bubble-client/styles"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	focusIndex int
	NameInput  textinput.Model
	Registered bool
	Active     bool
	Name       string
	width      int
	height     int
}

func InitialModel() Model {
	ni := textinput.New()
	ni.Placeholder = "Name"
	ni.Focus()
	ni.CharLimit = 156
	ni.Width = 20

	return Model{
		focusIndex: 0,
		NameInput:  ni,
		Registered: false,
		Active:     true,
		Name:       "",
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var nameInputCmd tea.Cmd

	m.NameInput, nameInputCmd = m.NameInput.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case messages.SetActive:
		m.Active = msg.Value
		if m.Active {
			m.NameInput.Blur()
		} else {
			m.NameInput.Focus()
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if !m.Active {
				return m, nil
			}
			if m.NameInput.Value() != "" {
				if m.focusIndex == 1 {
					m.Name = m.NameInput.Value()
					m.Active = false
					fmt.Println("should quit")
					return m, func() tea.Msg {
						return messages.NavigateTo{To: 1}
					}
				} else {
					m.NameInput.Blur()
					m.focusIndex = 1
				}
			}

		case tea.KeyTab:
			if m.focusIndex == 1 {
				m.NameInput.Focus()
				m.focusIndex = 0
			} else {
				m.NameInput.Blur()
				m.focusIndex = 1
			}
		}
	}

	return m, tea.Batch(nameInputCmd)
}

func (m Model) View() string {
	title := "Title title tiel"
	titleStyle := lipgloss.NewStyle().Bold(true).MarginBottom(1)
	buttonStyle := lipgloss.NewStyle().Bold(true).MarginTop(1)
	var button string
	if m.focusIndex == 0 {
		button = buttonStyle.Render(styles.BlurredButton)
	} else {
		button = buttonStyle.Render(styles.FocusedButton)
	}

	form := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.Gap,
		titleStyle.Render(title),
		m.NameInput.View(),
		button,
	)

	centered := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		form,
	)

	if !m.Active {
		return ""
	}

	return centered
}
