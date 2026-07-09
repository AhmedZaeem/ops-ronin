package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Search   key.Binding
	Admin    key.Binding
	Back     key.Binding
	Quit     key.Binding
	Yes      key.Binding
	No       key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Tab      key.Binding
	Pause    key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "run"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Admin: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "admin"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Yes: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "yes"),
	),
	No: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "no"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup", "shift+tab"),
		key.WithHelp("pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown", "tab"),
		key.WithHelp("pgdown", "page down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab", "right"),
		key.WithHelp("tab", "next tab"),
	),
	Pause: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "pause/resume"),
	),
}
