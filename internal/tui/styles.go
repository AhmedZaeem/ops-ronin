package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7AA2F7")).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#BB9AF7"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A9B1D6"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F7768E")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ECE6A"))

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0AF68"))

	containerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#73DACA")).
			Bold(true)

	viewportStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#565F89")).
			Padding(0, 1)

	dialogStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F7768E")).
			Padding(1, 2).
			Width(50)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565F89"))

	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7DCFFF"))

	logInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7AA2F7"))

	logWarnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0AF68"))

	logErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F7768E")).
			Bold(true)

	jsonKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BB9AF7"))

	jsonStringStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ECE6A"))

	jsonNumberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0AF68"))

	jsonBoolStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7DCFFF"))
)
