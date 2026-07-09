package tui

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/AhmedZaeem/ops-ronin/internal/docker"
)

func TestNewAdminModel(t *testing.T) {
	m := newAdminModel(100, 30)
	if !m.loading {
		t.Error("expected loading state")
	}
}

func TestLoadAdminData(t *testing.T) {
	exec := &docker.MockExecutor{
		Containers: []docker.ContainerSummary{{Name: "web", State: "running"}},
		Images:     []string{"nginx:latest"},
		Volumes:    []string{"data"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg := loadAdminData(ctx, exec)
	loaded, ok := msg.(adminLoadedMsg)
	if !ok {
		t.Fatalf("expected adminLoadedMsg")
	}
	if loaded.err != nil {
		t.Fatalf("unexpected error: %v", loaded.err)
	}
	if len(loaded.containers) != 1 {
		t.Errorf("expected 1 container, got %d", len(loaded.containers))
	}
}

func TestLoadAdminDataError(t *testing.T) {
	exec := &docker.MockExecutor{
		Containers: []docker.ContainerSummary{{Name: "fail", State: ""}},
	}
	// Force an error by making ListContainers return an error through a wrapper.
	failingExec := &failingListExecutor{MockExecutor: exec, err: errors.New("list failed")}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg := loadAdminData(ctx, failingExec)
	loaded := msg.(adminLoadedMsg)
	if loaded.err == nil {
		t.Fatal("expected error")
	}
}

func TestAdminModelUpdateData(t *testing.T) {
	m := newAdminModel(100, 30)
	m.UpdateData(adminLoadedMsg{
		containers: []docker.ContainerSummary{{Name: "web", State: "running"}},
		images:     []string{"nginx"},
		volumes:    []string{"data"},
	})

	if m.loading {
		t.Error("expected loading to be false")
	}
	if len(m.containers.Items()) != 1 {
		t.Errorf("expected 1 container item, got %d", len(m.containers.Items()))
	}
}

func TestModelOpensAdminPanel(t *testing.T) {
	cfg := testConfig()
	exec := &docker.MockExecutor{}
	m := modelInListState(t, cfg, exec)

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	if m.state != stateAdmin {
		t.Errorf("expected admin state, got %v", m.state)
	}
}

func TestAdminContainerActionStart(t *testing.T) {
	exec := &docker.MockExecutor{}
	m := modelInListState(t, testConfig(), exec)
	m.state = stateAdmin
	m.admin = newAdminModel(100, 30)
	m.admin.containers.SetItems([]list.Item{adminContainerItem{summary: docker.ContainerSummary{Name: "web", State: "exited"}}})

	_, cmd := m.runAdminAction()
	if cmd == nil {
		t.Fatal("expected command")
	}
}

func TestAdminPruneAction(t *testing.T) {
	exec := &docker.MockExecutor{PruneImagesOutput: "pruned"}
	m := modelInListState(t, testConfig(), exec)
	m.state = stateAdmin
	m.admin = newAdminModel(100, 30)
	m.admin.tab = adminTabPrune
	m.admin.prune.Select(1)

	msg := m.runAdminPruneAction(context.Background(), "Prune dangling images")
	if msg.err != nil {
		t.Fatalf("unexpected error: %v", msg.err)
	}
	if msg.output != "pruned" {
		t.Errorf("expected pruned, got %s", msg.output)
	}
}

// failingListExecutor wraps a MockExecutor and always fails ListContainers.
type failingListExecutor struct {
	*docker.MockExecutor
	err error
}

func (m *failingListExecutor) ListContainers(ctx context.Context) ([]docker.ContainerSummary, error) {
	return nil, m.err
}

var _ docker.Executor = (*failingListExecutor)(nil)

