package docker

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MockExecutor is a thread-safe in-memory implementation of Executor for tests.
type MockExecutor struct {
	mu sync.RWMutex

	PingErr error

	Containers []ContainerSummary
	Images     []string
	Volumes    []string
	Networks   []string

	ContainerLogsOutput   string
	ContainerLogsFollowCh <-chan LogChunk
	ContainerExecOutput   string
	ContainerStatusOutput string
	NetworkInspectOutput  string
	InfoOutput            string

	PruneContainersOutput string
	PruneImagesOutput     string
	PruneVolumesOutput    string
	PruneNetworksOutput   string

	StatsSamples []StatsSample
	StatsErr     error

	LastExecName    string
	LastExecCmd     []string
	LastStartedName string
	Calls           []string
}

func (m *MockExecutor) record(call string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, call)
}

// Ping verifies connectivity.
func (m *MockExecutor) Ping(ctx context.Context) error {
	m.record("Ping")
	return m.PingErr
}

// ContainerLogs returns recent logs.
func (m *MockExecutor) ContainerLogs(ctx context.Context, name string, tail string) (string, error) {
	m.record("ContainerLogs")
	return m.ContainerLogsOutput, nil
}

// ContainerLogsFollow returns the configured log channel.
func (m *MockExecutor) ContainerLogsFollow(ctx context.Context, name string, tail string) <-chan LogChunk {
	m.record("ContainerLogsFollow")
	if m.ContainerLogsFollowCh != nil {
		return m.ContainerLogsFollowCh
	}
	ch := make(chan LogChunk)
	close(ch)
	return ch
}

// ContainerExec runs a command inside a container.
func (m *MockExecutor) ContainerExec(ctx context.Context, name string, cmd []string) (string, error) {
	m.record("ContainerExec")
	m.LastExecName = name
	m.LastExecCmd = cmd
	return m.ContainerExecOutput, nil
}

// ContainerRestart restarts a container.
func (m *MockExecutor) ContainerRestart(ctx context.Context, name string) error {
	m.record("ContainerRestart")
	return nil
}

// ContainerStop stops a container.
func (m *MockExecutor) ContainerStop(ctx context.Context, name string) error {
	m.record("ContainerStop")
	return nil
}

// ContainerStart starts a container.
func (m *MockExecutor) ContainerStart(ctx context.Context, name string) error {
	m.record("ContainerStart")
	m.LastStartedName = name
	return nil
}

// ContainerStatus returns container inspect JSON.
func (m *MockExecutor) ContainerStatus(ctx context.Context, name string) (string, error) {
	m.record("ContainerStatus")
	return m.ContainerStatusOutput, nil
}

// ContainerRemove removes a container.
func (m *MockExecutor) ContainerRemove(ctx context.Context, name string) error {
	m.record("ContainerRemove")
	return nil
}

// ContainerStats streams resource utilization samples.
func (m *MockExecutor) ContainerStats(ctx context.Context, name string) (<-chan StatsSample, error) {
	m.record("ContainerStats")
	if m.StatsErr != nil {
		return nil, m.StatsErr
	}

	ch := make(chan StatsSample)
	go func() {
		defer close(ch)
		for _, s := range m.StatsSamples {
			select {
			case <-ctx.Done():
				return
			case ch <- s:
			}
		}
	}()

	return ch, nil
}

// PruneContainers removes stopped containers.
func (m *MockExecutor) PruneContainers(ctx context.Context) (string, error) {
	m.record("PruneContainers")
	return m.PruneContainersOutput, nil
}

// PruneImages removes dangling images.
func (m *MockExecutor) PruneImages(ctx context.Context) (string, error) {
	m.record("PruneImages")
	return m.PruneImagesOutput, nil
}

// PruneVolumes removes unused volumes.
func (m *MockExecutor) PruneVolumes(ctx context.Context) (string, error) {
	m.record("PruneVolumes")
	return m.PruneVolumesOutput, nil
}

// PruneNetworks removes unused networks.
func (m *MockExecutor) PruneNetworks(ctx context.Context) (string, error) {
	m.record("PruneNetworks")
	return m.PruneNetworksOutput, nil
}

// ListContainers returns all containers.
func (m *MockExecutor) ListContainers(ctx context.Context) ([]ContainerSummary, error) {
	m.record("ListContainers")
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]ContainerSummary{}, m.Containers...), nil
}

// ListImages returns image references.
func (m *MockExecutor) ListImages(ctx context.Context) ([]string, error) {
	m.record("ListImages")
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]string{}, m.Images...), nil
}

// ListVolumes returns volume references.
func (m *MockExecutor) ListVolumes(ctx context.Context) ([]string, error) {
	m.record("ListVolumes")
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]string{}, m.Volumes...), nil
}

// ListNetworks returns network references.
func (m *MockExecutor) ListNetworks(ctx context.Context) ([]string, error) {
	m.record("ListNetworks")
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]string{}, m.Networks...), nil
}

// InspectNetwork returns network JSON.
func (m *MockExecutor) InspectNetwork(ctx context.Context, id string) (string, error) {
	m.record("InspectNetwork")
	return m.NetworkInspectOutput, nil
}

// Info returns Docker daemon information.
func (m *MockExecutor) Info(ctx context.Context) (string, error) {
	m.record("Info")
	return m.InfoOutput, nil
}

var _ Executor = (*MockExecutor)(nil)

// NewMockLogChannel builds a channel that emits the provided chunks and closes.
func NewMockLogChannel(chunks []LogChunk) <-chan LogChunk {
	ch := make(chan LogChunk)
	go func() {
		defer close(ch)
		for _, c := range chunks {
			ch <- c
			time.Sleep(1 * time.Millisecond)
		}
	}()
	return ch
}

// NewMockStatsChannel builds a channel that emits the provided samples and closes.
func NewMockStatsChannel(samples []StatsSample) <-chan StatsSample {
	ch := make(chan StatsSample)
	go func() {
		defer close(ch)
		for _, s := range samples {
			ch <- s
			time.Sleep(1 * time.Millisecond)
		}
	}()
	return ch
}

// NewErrorMockExecutor returns a mock that fails every method with a fixed error.
func NewErrorMockExecutor() *MockExecutor {
	return &MockExecutor{
		PingErr:  fmt.Errorf("mock docker error"),
		StatsErr: fmt.Errorf("mock stats error"),
	}
}
