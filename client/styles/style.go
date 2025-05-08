package styles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	FocusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	BlurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	Gap          = "\n\n"

	FocusedButton = FocusedStyle.Render("[ Register ]")
	BlurredButton = fmt.Sprintf("[ %s ]", BlurredStyle.Render("Register"))

	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#e9d502"))
)
