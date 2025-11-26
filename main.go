package main

import (
	"fmt"
	"os"
	"strings"

	"ops-ronin/internal"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	red    = lipgloss.Color("#FF3333")
	green  = lipgloss.Color("#00AA00")
	yellow = lipgloss.Color("#FFAA00")
	grey   = lipgloss.Color("#555")
	white  = lipgloss.Color("#FFF")
	blue   = lipgloss.Color("#0088FF")

	selectedItemStyle = lipgloss.NewStyle().Foreground(white).Background(red).Bold(true).Padding(0, 1)
	itemStyle         = lipgloss.NewStyle().Foreground(grey).PaddingLeft(2)
	titleStyle        = lipgloss.NewStyle().Foreground(red).Bold(true).Border(lipgloss.DoubleBorder())
	errorStyle        = lipgloss.NewStyle().Foreground(white).Background(red).Padding(0, 1)
	successStyle      = lipgloss.NewStyle().Foreground(white).Background(green).Padding(0, 1)
	infoStyle         = lipgloss.NewStyle().Foreground(blue).Bold(true)
	helpStyle         = lipgloss.NewStyle().Foreground(grey).Italic(true)
)

type model struct {
	config     *internal.Config
	cursor     int
	flatTasks  []internal.Task
	lastOutput string
	executing  bool
	showHelp   bool
	error      string
	success    string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(" 🥷 Ops-Ronin - Universal TUI Engine ") + "\n\n")

	if m.config != nil {
		s.WriteString(infoStyle.Render(fmt.Sprintf("Project: %s", m.config.Project)) + "\n\n")
	}

	if len(m.flatTasks) == 0 {
		s.WriteString(errorStyle.Render(" No tasks found in menu.yaml ") + "\n\n")
		s.WriteString(helpStyle.Render("Add tasks to your menu.yaml file to get started!") + "\n")
	} else {
		for i, task := range m.flatTasks {
			cursor := " "
			style := itemStyle
			if m.cursor == i {
				cursor = ">"
				style = selectedItemStyle
			}
			s.WriteString(style.Render(cursor+" "+task.Label) + "\n")
		}
	}

	s.WriteString("\n" + helpStyle.Render("Controls:"))
	s.WriteString("\n" + helpStyle.Render("  ↑/k: Move up    ↓/j: Move down    Enter: Execute    h: Help    q: Quit"))

	if m.showHelp {
		s.WriteString("\n\n" + infoStyle.Render("📖 Help:"))
		s.WriteString("\n" + helpStyle.Render("• This tool executes commands in Docker containers"))
		s.WriteString("\n" + helpStyle.Render("• Commands are defined in menu.yaml"))
		s.WriteString("\n" + helpStyle.Render("• Make sure your containers are running before executing"))
		s.WriteString("\n" + helpStyle.Render("• Press 'h' again to hide this help"))
	}

	if m.executing {
		s.WriteString("\n\n" + infoStyle.Render("⏳ Executing command..."))
	}

	if m.error != "" {
		s.WriteString("\n\n" + errorStyle.Render("❌ Error: "+m.error))
	}

	if m.success != "" {
		s.WriteString("\n\n" + successStyle.Render("✅ "+m.success))
	}

	if m.lastOutput != "" {
		s.WriteString("\n\n" + infoStyle.Render("📋 Last Output:"))
		s.WriteString("\n" + lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).Render(m.lastOutput))
	}

	return s.String()
}

func initialModel() model {
	cfg, err := internal.LoadConfig("menu.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			return model{
				error: "menu.yaml not found. Please create a menu.yaml file with your tasks.",
			}
		}
		return model{
			error: fmt.Sprintf("Failed to load config: %v", err),
		}
	}

	var tasks []internal.Task
	for _, cat := range cfg.Menu {
		tasks = append(tasks, cat.Items...)
	}

	if len(tasks) == 0 {
		return model{
			config: cfg,
			error:  "No tasks found in menu.yaml. Add some tasks to get started!",
		}
	}

	return model{
		config:    cfg,
		flatTasks: tasks,
		success:   fmt.Sprintf("Loaded %d tasks successfully!", len(tasks)),
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.error = ""
		m.success = ""

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "h":
			m.showHelp = !m.showHelp
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.flatTasks)-1 {
				m.cursor++
			}
			return m, nil
		case "enter":
			if len(m.flatTasks) == 0 {
				m.error = "No tasks available to execute"
				return m, nil
			}

			m.executing = true
			selectedTask := m.flatTasks[m.cursor]

			m.lastOutput = ""

			output, err := internal.ExecuteCommand(selectedTask.Container, selectedTask.Command)
			m.executing = false

			if err != nil {
				m.error = err.Error()
				m.lastOutput = ""
			} else {
				m.success = fmt.Sprintf("Command executed successfully in %s", selectedTask.Container)
				m.lastOutput = output
			}
			return m, nil
		}
		return m, nil
	}
	return m, nil
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}

}
