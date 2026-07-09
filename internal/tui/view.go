package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/AhmedZaeem/ops-ronin/internal/monitoring"
)

// View renders the current TUI screen.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	header := m.renderHeader()

	switch m.state {
	case stateHealth:
		return lipgloss.JoinVertical(lipgloss.Left, header, m.renderHealth(), m.renderFooter())

	case stateAutoFix:
		return lipgloss.JoinVertical(lipgloss.Left, header, m.renderAutoFix(), m.renderFooter())

	case stateList:
		return lipgloss.JoinVertical(lipgloss.Left, header, m.list.View(), m.renderFooter())

	case stateSearch:
		return lipgloss.JoinVertical(lipgloss.Left, header, m.textInput.View(), m.list.View(), m.renderFooter())

	case stateConfirm:
		return lipgloss.JoinVertical(lipgloss.Left, header, m.renderConfirm(), m.renderFooter())

	case stateRunning:
		return lipgloss.JoinVertical(lipgloss.Left, header, m.renderRunning(), m.renderFooter())

	case stateOutput, stateError:
		return lipgloss.JoinVertical(lipgloss.Left, header, m.viewport.View(), m.renderFooter())

	case stateAdmin:
		return lipgloss.JoinVertical(lipgloss.Left, header, m.admin.View(m.width), m.renderFooter())

	case stateLogs:
		return lipgloss.JoinVertical(lipgloss.Left, header, m.renderLogs(), m.renderFooter())

	case stateMonitor:
		return lipgloss.JoinVertical(lipgloss.Left, header, m.renderMonitor(), m.renderFooter())

	default:
		return header + "\n"
	}
}

func (m *Model) renderHeader() string {
	return titleStyle.Render(m.cfg.Title)
}

func (m *Model) renderFooter() string {
	var parts []string

	switch m.state {
	case stateHealth:
		parts = []string{
			"enter: fix/continue",
			"↑/↓: select",
			"esc: skip",
		}
	case stateAutoFix:
		parts = []string{
			"enter: apply fix",
			"↑/↓: select suggestion",
			"esc: cancel",
		}
	case stateList:
		parts = []string{
			"enter: run",
			"/: search",
			"a: admin",
			"q: quit",
		}
	case stateSearch:
		parts = []string{
			"enter: apply",
			"esc: cancel",
		}
	case stateConfirm:
		parts = []string{
			"y: confirm",
			"n: cancel",
		}
	case stateOutput, stateError:
		parts = []string{
			"esc: back",
			"pgup/pgdn: scroll",
		}
	case stateRunning:
		parts = []string{"running..."}
	case stateAdmin:
		parts = []string{
			"tab: switch tab",
			"enter: action",
			"esc: back",
		}
	case stateLogs:
		status := "live"
		if m.logPaused {
			status = "paused"
		}
		parts = []string{
			fmt.Sprintf("status: %s", status),
			"p: pause/resume",
			"esc: back",
			"pgup/pgdn: scroll",
		}
	case stateMonitor:
		parts = []string{
			"esc: back",
		}
	}

	return helpStyle.Render(strings.Join(parts, "  •  "))
}

func (m *Model) renderHealth() string {
	if m.healthLoading {
		return lipgloss.Place(m.width, m.height-4, lipgloss.Center, lipgloss.Center,
			fmt.Sprintf("%s Checking container health...", m.spinner.View()))
	}

	if m.healthReport == nil {
		return "No health data"
	}

	var lines []string
	lines = append(lines, headerStyle.Render("Container Health"))
	lines = append(lines, "")

	allHealthy := true
	for i, entry := range m.healthReport.entries {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(entry.State.Color()))
		icon := style.Render(entry.State.Icon())
		name := entry.ConfigName
		if i == m.selectedHealthIndex {
			name = highlightStyle.Render("> " + name)
		} else {
			name = "  " + name
		}

		status := ""
		if entry.State == HealthMissing {
			status = "missing"
			allHealthy = false
		} else if entry.State == HealthStopped {
			status = entry.Status
			allHealthy = false
		} else {
			status = entry.Status
		}

		lines = append(lines, fmt.Sprintf("%s %s (%s)", icon, name, status))
	}

	lines = append(lines, "")
	if allHealthy {
		lines = append(lines, helpStyle.Render("All containers healthy. Press Enter to continue."))
	} else {
		lines = append(lines, helpStyle.Render("Select a container and press Enter to auto-fix."))
	}

	return lipgloss.Place(m.width, m.height-4, lipgloss.Center, lipgloss.Center,
		dialogStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...)))
}

func (m *Model) renderAutoFix() string {
	if m.healthReport == nil || m.selectedHealthIndex >= len(m.healthReport.entries) {
		return ""
	}

	entry := m.healthReport.entries[m.selectedHealthIndex]
	var lines []string
	lines = append(lines, headerStyle.Render("Auto Fix"))
	lines = append(lines, "")

	if entry.State == HealthStopped {
		lines = append(lines, fmt.Sprintf("Container '%s' is stopped.", entry.ConfigName))
		lines = append(lines, "")
		lines = append(lines, highlightStyle.Render("Press Enter to start it."))
	} else if entry.State == HealthMissing {
		lines = append(lines, fmt.Sprintf("Container '%s' was not found.", entry.ConfigName))
		lines = append(lines, "")

		if len(m.autoFixSuggestions) == 0 {
			lines = append(lines, "No similar containers found. Update menu.yaml and restart.")
		} else {
			lines = append(lines, "Similar containers found:")
			for i, s := range m.autoFixSuggestions {
				prefix := "  "
				if i == m.autoFixSelected {
					prefix = highlightStyle.Render("> ")
				}
				stateIcon := "●"
				if strings.ToLower(s.State) == "running" {
					stateIcon = successStyle.Render("●")
				} else {
					stateIcon = errorStyle.Render("●")
				}
				lines = append(lines, fmt.Sprintf("%s%s %s (%s)", prefix, stateIcon, s.Name, s.Status))
			}
			lines = append(lines, "")
			lines = append(lines, highlightStyle.Render("Press Enter to start/use selected container."))
		}
	}

	return lipgloss.Place(m.width, m.height-4, lipgloss.Center, lipgloss.Center,
		dialogStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...)))
}

func (m *Model) renderConfirm() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		headerStyle.Render("Confirm Action"),
		"",
		m.confirmMsg,
		"",
		highlightStyle.Render("Press y to confirm, n to cancel"),
	)
	return lipgloss.Place(m.width, m.height-4, lipgloss.Center, lipgloss.Center, dialogStyle.Render(content))
}

func (m *Model) renderRunning() string {
	var label string
	if m.state == stateAdmin {
		label = fmt.Sprintf("%s Running admin action...", m.spinner.View())
	} else {
		label = fmt.Sprintf("%s Running '%s' on '%s'...", m.spinner.View(), m.selectedItem.command.Label, m.selectedItem.container.Label)
	}
	return lipgloss.Place(m.width, m.height-4, lipgloss.Center, lipgloss.Center, label)
}

func (m *Model) renderError() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		errorStyle.Render("Error"),
		"",
		m.err.Error(),
		"",
		helpStyle.Render("Press esc to go back"),
	)
	return lipgloss.Place(m.width, m.height-4, lipgloss.Center, lipgloss.Center, dialogStyle.Render(content))
}

func (m *Model) renderLogs() string {
	status := "● LIVE"
	if m.logPaused {
		status = "● PAUSED"
	}

	headerLine := fmt.Sprintf("%s %s - %s",
		headerStyle.Render("Live Logs"),
		subtitleStyle.Render(m.selectedItem.container.Label),
		highlightStyle.Render(status),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		headerLine,
		m.viewport.View(),
	)
}

func (m *Model) renderMonitor() string {
	if m.cpuHistory.Len() == 0 {
		return lipgloss.Place(m.width, m.height-4, lipgloss.Center, lipgloss.Center,
			fmt.Sprintf("%s Collecting metrics for %s...", m.spinner.View(), m.selectedItem.container.Label))
	}

	latest := m.cpuHistory.Snapshot()[m.cpuHistory.Len()-1]

	var lines []string
	lines = append(lines, headerStyle.Render(fmt.Sprintf("Resource Monitor: %s", m.selectedItem.container.Label)))
	lines = append(lines, "")

	cpuBar := m.cpuProgress.ViewAs(latest.CPUUsage / 100.0)
	lines = append(lines, fmt.Sprintf("CPU  %.1f%%  %s", latest.CPUUsage, cpuBar))
	lines = append(lines, monitoring.SparklineWithLabel("CPU", m.cpuHistory.CPUValues(), latest.CPUUsage, "%"))
	lines = append(lines, "")

	memBar := m.memProgress.ViewAs(latest.MemoryPct / 100.0)
	lines = append(lines, fmt.Sprintf("Mem  %.1f%%  %s", latest.MemoryPct, memBar))
	lines = append(lines, monitoring.SparklineWithLabel("Mem", m.memHistory.MemoryValues(), latest.MemoryPct, "%"))
	lines = append(lines, fmt.Sprintf("      %s / %s",
		monitoring.FormatBytes(latest.MemoryUsage),
		monitoring.FormatBytes(latest.MemoryLimit)))
	lines = append(lines, "")

	lines = append(lines, subtitleStyle.Render(fmt.Sprintf("Network RX: %s  TX: %s",
		monitoring.FormatBytes(latest.NetworkRx),
		monitoring.FormatBytes(latest.NetworkTx))))

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.Place(m.width, m.height-4, lipgloss.Center, lipgloss.Center,
		dialogStyle.Width(m.width-8).Render(content))
}

// renderOutput applies lightweight JSON syntax highlighting when appropriate.
func renderOutput(output string) string {
	if output == "" {
		return "(no output)"
	}

	trimmed := strings.TrimSpace(output)
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
		return highlightJSON(trimmed)
	}

	return output
}

var (
	jsonKeyRegex    = regexp.MustCompile(`("[^"]+")\s*:`)
	jsonStringRegex = regexp.MustCompile(`:\s*("(?:[^"\\]|\\.)*")`)
	jsonNumberRegex = regexp.MustCompile(`:\s*(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)`)
	jsonBoolRegex   = regexp.MustCompile(`:\s*(true|false|null)`)
)

// highlightJSON applies simple regex-based syntax highlighting for JSON output.
func highlightJSON(input string) string {
	out := input
	out = jsonKeyRegex.ReplaceAllString(out, jsonKeyStyle.Render("$1")+":")
	out = jsonStringRegex.ReplaceAllString(out, ": "+jsonStringStyle.Render("$1"))
	out = jsonNumberRegex.ReplaceAllString(out, ": "+jsonNumberStyle.Render("$1"))
	out = jsonBoolRegex.ReplaceAllString(out, ": "+jsonBoolStyle.Render("$1"))
	return out
}
