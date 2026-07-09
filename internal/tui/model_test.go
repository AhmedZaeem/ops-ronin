package tui

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/AhmedZaeem/ops-ronin/internal/config"
	"github.com/AhmedZaeem/ops-ronin/internal/docker"
)

func modelInListState(t *testing.T, cfg *config.Config, exec docker.Executor) *Model {
	t.Helper()
	m := NewModel(cfg, exec)
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m.state = stateList
	return m
}

func testConfig() *config.Config {
	return &config.Config{
		Title: "Test",
		Containers: []config.Container{
			{
				Name:  "web",
				Label: "Web Server",
				Commands: []config.Command{
					{Name: "logs", Label: "Logs", Action: config.ActionLogs},
					{Name: "live-logs", Label: "Live Logs", Action: config.ActionLogsFollow},
					{Name: "monitor", Label: "Monitor", Action: config.ActionMonitor},
					{Name: "restart", Label: "Restart", Action: config.ActionRestart, Dangerous: true},
					{Name: "shell", Label: "Shell", Action: config.ActionExec, Command: "sh"},
				},
			},
		},
	}
}

func TestNewModel(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{}
	m := NewModel(cfg, exec)

	if m.cfg.Title != "Test" {
		t.Errorf("expected title Test, got %s", m.cfg.Title)
	}
	if m.state != stateHealth {
		t.Errorf("expected initial state health, got %v", m.state)
	}
}

func TestModelWindowSize(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{}
	m := NewModel(cfg, exec)

	m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height 40, got %d", m.height)
	}
}

func TestModelSelectDangerousCommandShowsConfirm(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{}
	m := modelInListState(t, cfg, exec)

	m.list.Select(3)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.state != stateConfirm {
		t.Errorf("expected confirm state, got %v", m.state)
	}
}

func TestModelConfirmRunsCommand(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{}
	m := modelInListState(t, cfg, exec)

	m.list.Select(3)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	if m.state != stateRunning {
		t.Errorf("expected running state, got %v", m.state)
	}
}

func TestModelCancelConfirm(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{}
	m := modelInListState(t, cfg, exec)

	m.list.Select(3)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	if m.state != stateList {
		t.Errorf("expected list state, got %v", m.state)
	}
}

func TestModelExecuteLogs(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{ContainerLogsOutput: "log line"}
	m := modelInListState(t, cfg, exec)

	m.list.Select(0)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.state != stateRunning {
		t.Errorf("expected running state, got %v", m.state)
	}
}

func TestModelRunFinished(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{ContainerLogsOutput: "log line"}
	m := modelInListState(t, cfg, exec)

	m.list.Select(0)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m.Update(runFinishedMsg{output: "log line"})

	if m.state != stateOutput {
		t.Errorf("expected output state, got %v", m.state)
	}
	if m.output != "log line" {
		t.Errorf("expected output log line, got %s", m.output)
	}
}

func TestModelRunError(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{ContainerLogsOutput: "log line"}
	m := modelInListState(t, cfg, exec)

	m.list.Select(0)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m.Update(runErrorMsg{err: errors.New("container not found")})

	if m.state != stateError {
		t.Errorf("expected error state, got %v", m.state)
	}
}

func TestModelSearchFilters(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{}
	m := modelInListState(t, cfg, exec)

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l', 'o', 'g', 's'}})

	if m.state != stateSearch {
		t.Errorf("expected search state, got %v", m.state)
	}
	if m.list.Items() == nil {
		t.Errorf("expected filtered items")
	}
}

func TestModelQuit(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{}
	m := NewModel(cfg, exec)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestBuildItems(t *testing.T) {
	cfg := testConfig()
	items := buildItems(cfg)
	if len(items) != 5 {
		t.Errorf("expected 5 items, got %d", len(items))
	}
}

func TestRenderOutput(t *testing.T) {
	if renderOutput("") != "(no output)" {
		t.Errorf("expected no output placeholder")
	}
	if renderOutput("hello") != "hello" {
		t.Errorf("expected hello")
	}
}

func TestRenderOutputJSON(t *testing.T) {
	out := renderOutput(`{"key":"value","num":42}`)
	if out == `{"key":"value","num":42}` {
		t.Error("expected JSON to be highlighted")
	}
}

func TestModelExecuteExec(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{ContainerExecOutput: "shell output"}
	m := modelInListState(t, cfg, exec)

	m.list.Select(4)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m.Update(runFinishedMsg{output: "shell output"})

	if m.state != stateOutput {
		t.Errorf("expected output state, got %v", m.state)
	}
	if m.output != "shell output" {
		t.Errorf("expected shell output, got %s", m.output)
	}
}

func TestModelOutputBackReturnsToList(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{ContainerLogsOutput: "log line"}
	m := modelInListState(t, cfg, exec)

	m.list.Select(0)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m.Update(runFinishedMsg{output: "log line"})
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if m.state != stateList {
		t.Errorf("expected list state, got %v", m.state)
	}
}

func TestModelSearchEscape(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{}
	m := modelInListState(t, cfg, exec)

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if m.state != stateList {
		t.Errorf("expected list state after escape, got %v", m.state)
	}
}

func TestModelQuitFromOutput(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{ContainerLogsOutput: "log line"}
	m := modelInListState(t, cfg, exec)

	m.list.Select(0)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m.Update(runFinishedMsg{output: "log line"})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestModelExecuteUnsupportedAction(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{}
	m := modelInListState(t, cfg, exec)

	m.selectedItem = commandItem{
		container: cfg.Containers[0],
		command:   config.Command{Name: "bad", Label: "Bad", Action: config.Action("invalid")},
	}

	_, err := m.execute(context.Background(), m.selectedItem)
	if err == nil {
		t.Fatal("expected error for unsupported action")
	}
}

func TestModelOpenLiveLogs(t *testing.T) {
	cfg := testConfig()
	chunks := []docker.LogChunk{
		{Stream: "stdout", Data: "log1\n"},
	}
	exec := &docker.MockExecutor{
		ContainerLogsFollowCh: docker.NewMockLogChannel(chunks),
	}
	m := modelInListState(t, cfg, exec)

	m.list.Select(1)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.state != stateLogs {
		t.Errorf("expected logs state, got %v", m.state)
	}
}

func TestModelOpenMonitor(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{
		StatsSamples: []docker.StatsSample{
			{CPUUsage: 10, MemoryPct: 20},
		},
	}
	m := modelInListState(t, cfg, exec)

	m.list.Select(2)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.state != stateMonitor {
		t.Errorf("expected monitor state, got %v", m.state)
	}
}

func TestModelLogChunkAppends(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{}
	m := modelInListState(t, cfg, exec)

	m.state = stateLogs
	m.appendLogChunk(docker.LogChunk{Stream: "stdout", Data: "hello\n"})

	if m.logBuffer == "" {
		t.Error("expected log buffer to contain rendered chunk")
	}
}
