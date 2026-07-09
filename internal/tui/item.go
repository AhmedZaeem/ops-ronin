package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"

	"github.com/AhmedZaeem/ops-ronin/internal/config"
)

// commandItem represents a single command entry in the main list.
type commandItem struct {
	container config.Container
	command   config.Command
}

// FilterValue is used by the fuzzy search/list filter.
func (i commandItem) FilterValue() string {
	return fmt.Sprintf("%s %s %s %s", i.container.Name, i.container.Label, i.command.Name, i.command.Label)
}

// Title returns the primary list text.
func (i commandItem) Title() string {
	return i.command.Label
}

// Description returns secondary list text.
func (i commandItem) Description() string {
	return fmt.Sprintf("%s  %s", i.container.Label, i.command.Name)
}

// buildItems flattens the configuration into list items.
func buildItems(cfg *config.Config) []list.Item {
	items := make([]list.Item, 0)
	for _, c := range cfg.Containers {
		for _, cmd := range c.Commands {
			items = append(items, commandItem{container: c, command: cmd})
		}
	}
	return items
}
