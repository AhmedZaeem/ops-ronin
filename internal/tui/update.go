package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/AhmedZaeem/ops-ronin/internal/config"
	"github.com/AhmedZaeem/ops-ronin/internal/docker"
	"github.com/AhmedZaeem/ops-ronin/internal/monitoring"
)

// Update handles all incoming messages and updates the model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case healthFinishedMsg:
		m.healthLoading = false
		if msg.err != nil {
			m.err = msg.err
			m.state = stateError
			return m, nil
		}
		m.healthReport = msg.report
		return m, nil

	case adminLoadedMsg:
		m.admin.UpdateData(msg)
		return m, nil

	case adminActionMsg:
		m.admin.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.state = stateError
			return m, nil
		}
		m.output = msg.output
		m.state = stateOutput
		m.viewport.SetContent(renderOutput(m.output))
		m.viewport.GotoTop()
		return m, nil

	case runFinishedMsg:
		m.output = msg.output
		m.state = stateOutput
		m.viewport.SetContent(renderOutput(m.output))
		m.viewport.GotoTop()
		return m, nil

	case runErrorMsg:
		m.err = msg.err
		m.state = stateError
		return m, nil

	case logChunkMsg:
		if m.state != stateLogs {
			return m, nil
		}
		if !m.logPaused {
			m.appendLogChunk(msg.chunk)
		}
		if m.logChannel != nil {
			return m, listenLogs(m.logChannel)
		}
		return m, nil

	case logDoneMsg:
		m.logChannel = nil
		if m.logCancelled != nil {
			m.logCancelled()
			m.logCancelled = nil
		}
		return m, nil

	case statsSampleMsg:
		if m.state != stateMonitor {
			return m, nil
		}
		m.cpuHistory.Push(msg.sample)
		m.memHistory.Push(msg.sample)
		if m.statsChannel != nil {
			return m, tea.Batch(
				listenStats(m.statsChannel),
				m.cpuProgress.SetPercent(msg.sample.CPUUsage/100.0),
				m.memProgress.SetPercent(msg.sample.MemoryPct/100.0),
			)
		}
		return m, nil

	case statsDoneMsg:
		m.statsChannel = nil
		if m.statsCancelled != nil {
			m.statsCancelled()
			m.statsCancelled = nil
		}
		return m, nil

	case progress.FrameMsg:
		newCPU, cmd := m.cpuProgress.Update(msg)
		m.cpuProgress = newCPU.(progress.Model)
		newMem, cmd2 := m.memProgress.Update(msg)
		m.memProgress = newMem.(progress.Model)
		return m, tea.Batch(cmd, cmd2)

	default:
		var cmds []tea.Cmd
		if m.state == stateRunning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		var listCmd tea.Cmd
		m.list, listCmd = m.list.Update(msg)
		cmds = append(cmds, listCmd)
		return m, tea.Batch(cmds...)
	}
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Quit) {
		m.cleanupStreams()
		return m, tea.Quit
	}

	switch m.state {
	case stateHealth:
		return m.handleHealthKey(msg)
	case stateAutoFix:
		return m.handleAutoFixKey(msg)
	case stateList:
		return m.handleListKey(msg)
	case stateSearch:
		return m.handleSearchKey(msg)
	case stateConfirm:
		return m.handleConfirmKey(msg)
	case stateOutput, stateError:
		return m.handleOutputKey(msg)
	case stateRunning:
		return m, nil
	case stateAdmin:
		return m.handleAdminKey(msg)
	case stateLogs:
		return m.handleLogsKey(msg)
	case stateMonitor:
		return m.handleMonitorKey(msg)
	}

	return m, nil
}

func (m *Model) handleHealthKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Enter):
		if m.healthReport == nil {
			return m, nil
		}
		if m.selectedHealthIndex >= 0 && m.selectedHealthIndex < len(m.healthReport.entries) {
			entry := m.healthReport.entries[m.selectedHealthIndex]
			if entry.State == HealthStopped || entry.State == HealthMissing {
				return m.openAutoFix(entry)
			}
		}
		m.state = stateList
		return m, nil

	case key.Matches(msg, keys.Down):
		if m.healthReport != nil && m.selectedHealthIndex < len(m.healthReport.entries)-1 {
			m.selectedHealthIndex++
		}
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.selectedHealthIndex > 0 {
			m.selectedHealthIndex--
		}
		return m, nil

	case key.Matches(msg, keys.Back):
		m.state = stateList
		return m, nil
	}

	return m, nil
}

func (m *Model) openAutoFix(entry healthEntry) (tea.Model, tea.Cmd) {
	m.autoFixSelected = 0
	m.autoFixSuggestions = nil

	if entry.State == HealthMissing {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		containers, err := m.executor.ListContainers(ctx)
		if err == nil {
			m.autoFixSuggestions = findSuggestions(containers, entry.ConfigName)
		}
	}

	m.state = stateAutoFix
	return m, nil
}

func (m *Model) handleAutoFixKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	entry := m.healthReport.entries[m.selectedHealthIndex]

	switch {
	case key.Matches(msg, keys.Enter):
		return m.applyAutoFix(entry)

	case key.Matches(msg, keys.Back), key.Matches(msg, keys.No):
		m.state = stateHealth
		return m, nil

	case key.Matches(msg, keys.Down):
		if entry.State == HealthMissing && m.autoFixSelected < len(m.autoFixSuggestions)-1 {
			m.autoFixSelected++
		}
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.autoFixSelected > 0 {
			m.autoFixSelected--
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) applyAutoFix(entry healthEntry) (tea.Model, tea.Cmd) {
	m.state = stateRunning

	cmd := func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if entry.State == HealthMissing && len(m.autoFixSuggestions) > 0 {
			suggestion := m.autoFixSuggestions[m.autoFixSelected]
			if strings.ToLower(suggestion.State) != "running" {
				if err := m.executor.ContainerStart(ctx, suggestion.Name); err != nil {
					return runErrorMsg{err: fmt.Errorf("start suggested container: %w", err)}
				}
			}
			return runFinishedMsg{output: fmt.Sprintf("Using container '%s' for '%s'", suggestion.Name, entry.ConfigName)}
		}

		out, err := applyFix(ctx, m.executor, entry)
		if err != nil {
			return runErrorMsg{err: err}
		}
		return runFinishedMsg{output: out}
	}

	return m, tea.Batch(cmd, m.spinner.Tick)
}

func (m *Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Search):
		m.state = stateSearch
		m.textInput.Focus()
		return m, textinput.Blink

	case key.Matches(msg, keys.Admin):
		m.state = stateAdmin
		m.admin = newAdminModel(m.width, m.height-6)
		return m, func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			return loadAdminData(ctx, m.executor)
		}

	case key.Matches(msg, keys.Enter):
		if item, ok := m.list.SelectedItem().(commandItem); ok {
			m.selectedItem = item
			if item.command.IsDangerous() {
				m.confirmMsg = fmt.Sprintf("Run '%s' on '%s'?", item.command.Label, item.container.Label)
				m.state = stateConfirm
				return m, nil
			}
			return m.runSelected()
		}

	default:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.state = stateList
		m.textInput.SetValue("")
		m.list.ResetFilter()
		return m, nil

	case key.Matches(msg, keys.Enter):
		m.state = stateList
		m.list.ResetFilter()
		m.applySearch(m.textInput.Value())
		return m, nil

	default:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		m.applySearch(m.textInput.Value())
		return m, cmd
	}
}

func (m *Model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Yes):
		return m.runSelected()

	case key.Matches(msg, keys.No), key.Matches(msg, keys.Back):
		m.state = stateList
		return m, nil
	}

	return m, nil
}

func (m *Model) handleOutputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.state = stateList
		m.output = ""
		m.err = nil
		return m, nil

	case key.Matches(msg, keys.PageUp):
		m.viewport.LineUp(3)
		return m, nil

	case key.Matches(msg, keys.PageDown):
		m.viewport.LineDown(3)
		return m, nil

	default:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
}

func (m *Model) handleLogsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.cleanupStreams()
		m.state = stateList
		m.logBuffer = ""
		return m, nil

	case key.Matches(msg, keys.Pause):
		m.logPaused = !m.logPaused
		return m, nil

	case key.Matches(msg, keys.PageUp):
		m.viewport.LineUp(3)
		return m, nil

	case key.Matches(msg, keys.PageDown):
		m.viewport.LineDown(3)
		return m, nil

	default:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
}

func (m *Model) handleMonitorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.cleanupStreams()
		m.state = stateList
		m.cpuHistory = monitoring.NewRingBuffer(60)
		m.memHistory = monitoring.NewRingBuffer(60)
		return m, nil

	default:
		return m, nil
	}
}

func (m *Model) handleAdminKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.state = stateList
		return m, nil

	case key.Matches(msg, keys.Tab):
		m.admin.tab = (m.admin.tab + 1) % 4
		return m, nil

	case key.Matches(msg, keys.Enter):
		return m.runAdminAction()

	case key.Matches(msg, keys.Down), key.Matches(msg, keys.Up):
		lst := m.admin.CurrentList()
		var cmd tea.Cmd
		*lst, cmd = lst.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) runAdminAction() (tea.Model, tea.Cmd) {
	m.admin.loading = true

	cmd := func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		switch m.admin.tab {
		case adminTabContainers:
			if item, ok := m.admin.containers.SelectedItem().(adminContainerItem); ok {
				return m.runAdminContainerAction(ctx, item.summary)
			}

		case adminTabPrune:
			if item, ok := m.admin.prune.SelectedItem().(adminStringItem); ok {
				return m.runAdminPruneAction(ctx, item.value)
			}
		}

		return adminActionMsg{err: fmt.Errorf("no action selected")}
	}

	return m, tea.Batch(cmd, m.spinner.Tick)
}

func (m *Model) runAdminContainerAction(ctx context.Context, summary docker.ContainerSummary) adminActionMsg {
	state := strings.ToLower(summary.State)
	if state == "running" {
		if err := m.executor.ContainerStop(ctx, summary.Name); err != nil {
			return adminActionMsg{err: err}
		}
		return adminActionMsg{output: fmt.Sprintf("Container '%s' stopped", summary.Name)}
	}

	if err := m.executor.ContainerStart(ctx, summary.Name); err != nil {
		return adminActionMsg{err: err}
	}
	return adminActionMsg{output: fmt.Sprintf("Container '%s' started", summary.Name)}
}

func (m *Model) runAdminPruneAction(ctx context.Context, label string) adminActionMsg {
	var out string
	var err error

	switch label {
	case "Prune stopped containers":
		out, err = m.executor.PruneContainers(ctx)
	case "Prune dangling images":
		out, err = m.executor.PruneImages(ctx)
	case "Prune unused volumes":
		out, err = m.executor.PruneVolumes(ctx)
	case "Prune unused networks":
		out, err = m.executor.PruneNetworks(ctx)
	default:
		return adminActionMsg{err: fmt.Errorf("unknown prune action: %s", label)}
	}

	if err != nil {
		return adminActionMsg{err: err}
	}
	return adminActionMsg{output: out}
}

func (m *Model) applySearch(query string) {
	if query == "" {
		m.list.ResetFilter()
		return
	}

	filtered := make([]list.Item, 0)
	lower := strings.ToLower(query)
	for _, item := range buildItems(m.cfg) {
		if strings.Contains(strings.ToLower(item.FilterValue()), lower) {
			filtered = append(filtered, item)
		}
	}

	m.list.SetItems(filtered)
}

func (m *Model) runSelected() (tea.Model, tea.Cmd) {
	m.output = ""
	m.err = nil

	switch m.selectedItem.command.Action {
	case config.ActionLogsFollow:
		return m.runLogsFollow()
	case config.ActionStats, config.ActionMonitor:
		return m.runMonitor()
	case config.ActionImages, config.ActionVolumes, config.ActionNetworks:
		return m.runSystemList()
	case config.ActionNetworkInspect:
		return m.runNetworkInspect()
	default:
		return m.runOneShot()
	}
}

func (m *Model) runOneShot() (tea.Model, tea.Cmd) {
	m.state = stateRunning

	cmd := func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		out, err := m.execute(ctx, m.selectedItem)
		if err != nil {
			return runErrorMsg{err: err}
		}
		return runFinishedMsg{output: out}
	}

	return m, tea.Batch(cmd, m.spinner.Tick)
}

func (m *Model) runLogsFollow() (tea.Model, tea.Cmd) {
	m.cleanupStreams()
	m.state = stateLogs
	m.logBuffer = ""
	m.logPaused = false
	m.viewport.SetContent("")
	m.viewport.GotoBottom()

	ctx, cancel := context.WithCancel(context.Background())
	m.logCancelled = cancel

	name := m.selectedItem.container.Name
	tail := "100"
	if len(m.selectedItem.command.Args) >= 2 && m.selectedItem.command.Args[0] == "--tail" {
		tail = m.selectedItem.command.Args[1]
	}

	ch := m.executor.ContainerLogsFollow(ctx, name, tail)
	m.logChannel = ch

	return m, tea.Batch(
		listenLogs(ch),
		m.spinner.Tick,
	)
}

func (m *Model) runMonitor() (tea.Model, tea.Cmd) {
	m.cleanupStreams()
	m.state = stateMonitor
	m.cpuHistory = monitoring.NewRingBuffer(60)
	m.memHistory = monitoring.NewRingBuffer(60)

	ctx, cancel := context.WithCancel(context.Background())
	m.statsCancelled = cancel

	name := m.selectedItem.container.Name
	ch, err := m.executor.ContainerStats(ctx, name)
	if err != nil {
		m.err = err
		m.state = stateError
		return m, nil
	}
	m.statsChannel = ch

	return m, tea.Batch(
		listenStats(ch),
		m.spinner.Tick,
	)
}

func (m *Model) runSystemList() (tea.Model, tea.Cmd) {
	m.state = stateRunning

	cmd := func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var lines []string
		var err error

		switch m.selectedItem.command.Action {
		case config.ActionImages:
			lines, err = m.executor.ListImages(ctx)
		case config.ActionVolumes:
			lines, err = m.executor.ListVolumes(ctx)
		case config.ActionNetworks:
			lines, err = m.executor.ListNetworks(ctx)
		}

		if err != nil {
			return runErrorMsg{err: err}
		}

		if len(lines) == 0 {
			return runFinishedMsg{output: "No items found"}
		}
		return runFinishedMsg{output: strings.Join(lines, "\n")}
	}

	return m, tea.Batch(cmd, m.spinner.Tick)
}

func (m *Model) runNetworkInspect() (tea.Model, tea.Cmd) {
	m.state = stateRunning

	cmd := func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		id := m.selectedItem.container.Name
		if len(m.selectedItem.command.Args) > 0 {
			id = m.selectedItem.command.Args[0]
		}

		out, err := m.executor.InspectNetwork(ctx, id)
		if err != nil {
			return runErrorMsg{err: err}
		}
		return runFinishedMsg{output: out}
	}

	return m, tea.Batch(cmd, m.spinner.Tick)
}

func (m *Model) execute(ctx context.Context, item commandItem) (string, error) {
	name := item.container.Name

	switch item.command.Action {
	case config.ActionLogs:
		tail := "100"
		if len(item.command.Args) >= 2 && item.command.Args[0] == "--tail" {
			tail = item.command.Args[1]
		}
		return m.executor.ContainerLogs(ctx, name, tail)

	case config.ActionExec:
		cmd := []string{item.command.Command}
		cmd = append(cmd, item.command.Args...)
		return m.executor.ContainerExec(ctx, name, cmd)

	case config.ActionRestart:
		if err := m.executor.ContainerRestart(ctx, name); err != nil {
			return "", err
		}
		return fmt.Sprintf("Container '%s' restarted successfully", name), nil

	case config.ActionStop:
		if err := m.executor.ContainerStop(ctx, name); err != nil {
			return "", err
		}
		return fmt.Sprintf("Container '%s' stopped successfully", name), nil

	case config.ActionStart:
		if err := m.executor.ContainerStart(ctx, name); err != nil {
			return "", err
		}
		return fmt.Sprintf("Container '%s' started successfully", name), nil

	case config.ActionStatus:
		return m.executor.ContainerStatus(ctx, name)

	case config.ActionRemove:
		if err := m.executor.ContainerRemove(ctx, name); err != nil {
			return "", err
		}
		return fmt.Sprintf("Container '%s' removed successfully", name), nil

	case config.ActionPrune:
		return m.executor.PruneContainers(ctx)

	case config.ActionStats:
		// stats action defaults to one-shot snapshot; monitor is the streaming variant.
		return "Use the 'monitor' action for live stats, or run 'docker stats' from exec.", nil

	default:
		return "", fmt.Errorf("unsupported action: %s", item.command.Action)
	}
}

func (m *Model) appendLogChunk(chunk docker.LogChunk) {
	style := logInfoStyle
	lower := strings.ToLower(chunk.Data)
	if strings.Contains(lower, "error") || strings.Contains(lower, "fatal") || strings.Contains(lower, "panic") {
		style = logErrorStyle
	} else if strings.Contains(lower, "warn") {
		style = logWarnStyle
	} else if strings.Contains(lower, "info") {
		style = logInfoStyle
	}

	m.logBuffer += style.Render(chunk.Data)
	// Maintain a bounded buffer to avoid unbounded memory growth during long streams.
	const maxBufferLen = 256 * 1024
	if len(m.logBuffer) > maxBufferLen {
		m.logBuffer = m.logBuffer[len(m.logBuffer)-maxBufferLen:]
	}
	m.viewport.SetContent(m.logBuffer)
	m.viewport.GotoBottom()
}

func (m *Model) cleanupStreams() {
	if m.logCancelled != nil {
		m.logCancelled()
		m.logCancelled = nil
	}
	if m.statsCancelled != nil {
		m.statsCancelled()
		m.statsCancelled = nil
	}
	m.logChannel = nil
	m.statsChannel = nil
	m.logPaused = false
}

// listenLogs returns a Cmd that waits for the next log chunk.
func listenLogs(ch <-chan docker.LogChunk) tea.Cmd {
	return func() tea.Msg {
		chunk, ok := <-ch
		if !ok {
			return logDoneMsg{}
		}
		return logChunkMsg{chunk: chunk}
	}
}

// listenStats returns a Cmd that waits for the next stats sample.
func listenStats(ch <-chan docker.StatsSample) tea.Cmd {
	return func() tea.Msg {
		sample, ok := <-ch
		if !ok {
			return statsDoneMsg{}
		}
		return statsSampleMsg{sample: sample}
	}
}
