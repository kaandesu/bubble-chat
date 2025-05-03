package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	go Connect(serverAddr, p)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
