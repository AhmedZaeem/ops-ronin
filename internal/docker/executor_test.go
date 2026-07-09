package docker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryExecutorSuccess(t *testing.T) {
	mock := &MockExecutor{ContainerLogsOutput: "hello"}
	retry := NewRetryExecutor(mock)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := retry.ContainerLogs(ctx, "web", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out != "hello" {
		t.Errorf("expected hello, got %s", out)
	}
}

func TestRetryExecutorRetriesThenSucceeds(t *testing.T) {
	mock := &MockExecutor{
		ContainerLogsOutput: "hello",
	}
	retry := NewRetryExecutor(mock)
	retry.inner = &failThenSucceedExecutor{
		failures: 2,
		success:  mock,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := retry.ContainerLogs(ctx, "web", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out != "hello" {
		t.Errorf("expected hello, got %s", out)
	}
}

func TestRetryExecutorMaxRetries(t *testing.T) {
	mock := &MockExecutor{PingErr: errors.New("connection refused")}
	retry := NewRetryExecutor(mock)
	retry.inner = &alwaysFailExecutor{err: errors.New("connection refused")}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := retry.ContainerLogs(ctx, "web", "")
	if err == nil {
		t.Fatal("expected error after max retries")
	}
	if !errors.Is(err, ErrMaxRetries) {
		t.Errorf("expected ErrMaxRetries, got %v", err)
	}
}

func TestRetryExecutorNonRetryable(t *testing.T) {
	retry := NewRetryExecutor(&alwaysFailExecutor{err: errors.New("container not found")})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := retry.ContainerLogs(ctx, "web", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRetryExecutorContextCancel(t *testing.T) {
	retry := NewRetryExecutor(&alwaysFailExecutor{err: errors.New("connection refused")})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := retry.ContainerLogs(ctx, "web", "")
	if err == nil {
		t.Fatal("expected context canceled error")
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"connection refused", errors.New("connection refused"), true},
		{"i/o timeout", errors.New("i/o timeout"), true},
		{"context deadline exceeded", errors.New("context deadline exceeded"), true},
		{"container not found", errors.New("container not found"), false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryable(tt.err); got != tt.want {
				t.Errorf("isRetryable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// failThenSucceedExecutor fails the first N calls then delegates to success.
type failThenSucceedExecutor struct {
	failures int
	called   int
	success  Executor
}

func (m *failThenSucceedExecutor) Ping(ctx context.Context) error { return nil }
func (m *failThenSucceedExecutor) ContainerLogs(ctx context.Context, name string, tail string) (string, error) {
	m.called++
	if m.called <= m.failures {
		return "", errors.New("connection refused")
	}
	return m.success.ContainerLogs(ctx, name, tail)
}
func (m *failThenSucceedExecutor) ContainerLogsFollow(ctx context.Context, name string, tail string) <-chan LogChunk {
	return nil
}
func (m *failThenSucceedExecutor) ContainerExec(ctx context.Context, name string, cmd []string) (string, error) {
	return "", nil
}
func (m *failThenSucceedExecutor) ContainerRestart(ctx context.Context, name string) error { return nil }
func (m *failThenSucceedExecutor) ContainerStop(ctx context.Context, name string) error    { return nil }
func (m *failThenSucceedExecutor) ContainerStart(ctx context.Context, name string) error   { return nil }
func (m *failThenSucceedExecutor) ContainerStatus(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (m *failThenSucceedExecutor) ContainerRemove(ctx context.Context, name string) error { return nil }
func (m *failThenSucceedExecutor) ContainerStats(ctx context.Context, name string) (<-chan StatsSample, error) {
	return nil, nil
}
func (m *failThenSucceedExecutor) PruneContainers(ctx context.Context) (string, error) { return "", nil }
func (m *failThenSucceedExecutor) PruneImages(ctx context.Context) (string, error)     { return "", nil }
func (m *failThenSucceedExecutor) PruneVolumes(ctx context.Context) (string, error)    { return "", nil }
func (m *failThenSucceedExecutor) PruneNetworks(ctx context.Context) (string, error)   { return "", nil }
func (m *failThenSucceedExecutor) ListContainers(ctx context.Context) ([]ContainerSummary, error) {
	return nil, nil
}
func (m *failThenSucceedExecutor) ListImages(ctx context.Context) ([]string, error)   { return nil, nil }
func (m *failThenSucceedExecutor) ListVolumes(ctx context.Context) ([]string, error)   { return nil, nil }
func (m *failThenSucceedExecutor) ListNetworks(ctx context.Context) ([]string, error)  { return nil, nil }
func (m *failThenSucceedExecutor) InspectNetwork(ctx context.Context, id string) (string, error) {
	return "", nil
}
func (m *failThenSucceedExecutor) Info(ctx context.Context) (string, error) { return "", nil }

// alwaysFailExecutor always returns a fixed error.
type alwaysFailExecutor struct {
	err error
}

func (m *alwaysFailExecutor) Ping(ctx context.Context) error { return m.err }
func (m *alwaysFailExecutor) ContainerLogs(ctx context.Context, name string, tail string) (string, error) {
	return "", m.err
}
func (m *alwaysFailExecutor) ContainerLogsFollow(ctx context.Context, name string, tail string) <-chan LogChunk {
	return nil
}
func (m *alwaysFailExecutor) ContainerExec(ctx context.Context, name string, cmd []string) (string, error) {
	return "", m.err
}
func (m *alwaysFailExecutor) ContainerRestart(ctx context.Context, name string) error { return m.err }
func (m *alwaysFailExecutor) ContainerStop(ctx context.Context, name string) error    { return m.err }
func (m *alwaysFailExecutor) ContainerStart(ctx context.Context, name string) error   { return m.err }
func (m *alwaysFailExecutor) ContainerStatus(ctx context.Context, name string) (string, error) {
	return "", m.err
}
func (m *alwaysFailExecutor) ContainerRemove(ctx context.Context, name string) error { return m.err }
func (m *alwaysFailExecutor) ContainerStats(ctx context.Context, name string) (<-chan StatsSample, error) {
	return nil, m.err
}
func (m *alwaysFailExecutor) PruneContainers(ctx context.Context) (string, error) { return "", m.err }
func (m *alwaysFailExecutor) PruneImages(ctx context.Context) (string, error)     { return "", m.err }
func (m *alwaysFailExecutor) PruneVolumes(ctx context.Context) (string, error)    { return "", m.err }
func (m *alwaysFailExecutor) PruneNetworks(ctx context.Context) (string, error)   { return "", m.err }
func (m *alwaysFailExecutor) ListContainers(ctx context.Context) ([]ContainerSummary, error) {
	return nil, m.err
}
func (m *alwaysFailExecutor) ListImages(ctx context.Context) ([]string, error)   { return nil, m.err }
func (m *alwaysFailExecutor) ListVolumes(ctx context.Context) ([]string, error)   { return nil, m.err }
func (m *alwaysFailExecutor) ListNetworks(ctx context.Context) ([]string, error)  { return nil, m.err }
func (m *alwaysFailExecutor) InspectNetwork(ctx context.Context, id string) (string, error) {
	return "", m.err
}
func (m *alwaysFailExecutor) Info(ctx context.Context) (string, error) { return "", m.err }
