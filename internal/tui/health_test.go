package tui

import (
	"context"
	"errors"
	"testing"

	"github.com/AhmedZaeem/ops-ronin/internal/docker"
)

func TestCheckContainerHealthAllRunning(t *testing.T) {
	exec := &docker.MockExecutor{
		Containers: []docker.ContainerSummary{
			{Name: "web", State: "running", Status: "Up 2 hours"},
		},
	}

	report, err := checkContainerHealth(context.Background(), exec, []string{"web"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(report.entries))
	}
	if report.entries[0].State != HealthRunning {
		t.Errorf("expected running, got %v", report.entries[0].State)
	}
}

func TestCheckContainerHealthMissing(t *testing.T) {
	exec := &docker.MockExecutor{
		Containers: []docker.ContainerSummary{
			{Name: "web", State: "running"},
		},
	}

	report, err := checkContainerHealth(context.Background(), exec, []string{"web", "db"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.entries[1].State != HealthMissing {
		t.Errorf("expected missing, got %v", report.entries[1].State)
	}
}

func TestCheckContainerHealthStopped(t *testing.T) {
	exec := &docker.MockExecutor{
		Containers: []docker.ContainerSummary{
			{Name: "web", State: "exited", Status: "Exited (0)"},
		},
	}

	report, err := checkContainerHealth(context.Background(), exec, []string{"web"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.entries[0].State != HealthStopped {
		t.Errorf("expected stopped, got %v", report.entries[0].State)
	}
}

func TestCheckContainerHealthListError(t *testing.T) {
	exec := &healthFailingExecutor{err: errors.New("daemon unreachable")}
	_, err := checkContainerHealth(context.Background(), exec, []string{"web"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFindSuggestions(t *testing.T) {
	available := []docker.ContainerSummary{
		{Name: "nginx-proxy"},
		{Name: "postgres-db"},
		{Name: "redis-cache"},
	}

	suggestions := findSuggestions(available, "nginx")
	if len(suggestions) != 1 {
		t.Errorf("expected 1 suggestion, got %d", len(suggestions))
	}
	if suggestions[0].Name != "nginx-proxy" {
		t.Errorf("expected nginx-proxy, got %s", suggestions[0].Name)
	}
}

func TestApplyFixStartStopped(t *testing.T) {
	exec := &docker.MockExecutor{}
	entry := healthEntry{State: HealthStopped, ActualName: "web"}

	out, err := applyFix(context.Background(), exec, entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.LastStartedName != "web" {
		t.Errorf("expected web to be started, got %s", exec.LastStartedName)
	}
	if out == "" {
		t.Errorf("expected output")
	}
}

func TestApplyFixMissingNoFix(t *testing.T) {
	exec := &docker.MockExecutor{}
	entry := healthEntry{State: HealthMissing, ConfigName: "web"}

	_, err := applyFix(context.Background(), exec, entry)
	if err == nil {
		t.Fatal("expected error for missing container")
	}
}

// healthFailingExecutor wraps a MockExecutor and always fails ListContainers.
type healthFailingExecutor struct {
	*docker.MockExecutor
	err error
}

func (m *healthFailingExecutor) ListContainers(ctx context.Context) ([]docker.ContainerSummary, error) {
	return nil, m.err
}

var _ docker.Executor = (*healthFailingExecutor)(nil)
