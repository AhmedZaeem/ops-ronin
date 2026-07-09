package docker

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrMaxRetries = errors.New("max retries exceeded")

const maxRetries = 3

func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	text := strings.ToLower(err.Error())
	retryable := []string{
		"connection refused",
		"no such host",
		"i/o timeout",
		"broken pipe",
		"reset by peer",
		"temporary",
		"context deadline exceeded",
	}

	for _, marker := range retryable {
		if strings.Contains(text, marker) {
			return true
		}
	}

	return false
}

// RetryExecutor wraps an Executor with exponential backoff for transient errors.
type RetryExecutor struct {
	inner  Executor
	policy BackoffPolicy
}

// BackoffPolicy configures retry delays.
type BackoffPolicy struct {
	Base   time.Duration
	Factor float64
	Max    time.Duration
}

// DefaultBackoffPolicy returns a sensible retry configuration.
func DefaultBackoffPolicy() BackoffPolicy {
	return BackoffPolicy{
		Base:   500 * time.Millisecond,
		Factor: 2.0,
		Max:    5 * time.Second,
	}
}

// NewRetryExecutor creates a retrying decorator around the provided executor.
func NewRetryExecutor(inner Executor) *RetryExecutor {
	return &RetryExecutor{
		inner:  inner,
		policy: DefaultBackoffPolicy(),
	}
}

func (r *RetryExecutor) sleep(ctx context.Context, attempt int) error {
	delay := r.policy.Base
	for i := 0; i < attempt; i++ {
		delay = time.Duration(float64(delay) * r.policy.Factor)
		if delay > r.policy.Max {
			delay = r.policy.Max
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}

// Ping verifies connectivity to the Docker daemon.
func (r *RetryExecutor) Ping(ctx context.Context) error {
	return r.inner.Ping(ctx)
}

// ContainerLogs returns the most recent logs for a container.
func (r *RetryExecutor) ContainerLogs(ctx context.Context, name string, tail string) (string, error) {
	return retry(ctx, r, func() (string, error) {
		return r.inner.ContainerLogs(ctx, name, tail)
	})
}

// ContainerLogsFollow streams container logs; streaming methods are not retried.
func (r *RetryExecutor) ContainerLogsFollow(ctx context.Context, name string, tail string) <-chan LogChunk {
	return r.inner.ContainerLogsFollow(ctx, name, tail)
}

// ContainerExec runs a command inside a container.
func (r *RetryExecutor) ContainerExec(ctx context.Context, name string, cmd []string) (string, error) {
	return retry(ctx, r, func() (string, error) {
		return r.inner.ContainerExec(ctx, name, cmd)
	})
}

// ContainerRestart restarts a container.
func (r *RetryExecutor) ContainerRestart(ctx context.Context, name string) error {
	_, err := retry(ctx, r, func() (string, error) {
		return "", r.inner.ContainerRestart(ctx, name)
	})
	return err
}

// ContainerStop stops a container.
func (r *RetryExecutor) ContainerStop(ctx context.Context, name string) error {
	_, err := retry(ctx, r, func() (string, error) {
		return "", r.inner.ContainerStop(ctx, name)
	})
	return err
}

// ContainerStart starts a container.
func (r *RetryExecutor) ContainerStart(ctx context.Context, name string) error {
	_, err := retry(ctx, r, func() (string, error) {
		return "", r.inner.ContainerStart(ctx, name)
	})
	return err
}

// ContainerStatus returns container inspect JSON.
func (r *RetryExecutor) ContainerStatus(ctx context.Context, name string) (string, error) {
	return retry(ctx, r, func() (string, error) {
		return r.inner.ContainerStatus(ctx, name)
	})
}

// ContainerRemove removes a container.
func (r *RetryExecutor) ContainerRemove(ctx context.Context, name string) error {
	_, err := retry(ctx, r, func() (string, error) {
		return "", r.inner.ContainerRemove(ctx, name)
	})
	return err
}

// ContainerStats streams resource utilization samples; streaming methods are not retried.
func (r *RetryExecutor) ContainerStats(ctx context.Context, name string) (<-chan StatsSample, error) {
	return r.inner.ContainerStats(ctx, name)
}

// PruneContainers removes stopped containers.
func (r *RetryExecutor) PruneContainers(ctx context.Context) (string, error) {
	return retry(ctx, r, func() (string, error) {
		return r.inner.PruneContainers(ctx)
	})
}

// ListContainers returns all containers.
func (r *RetryExecutor) ListContainers(ctx context.Context) ([]ContainerSummary, error) {
	return retrySlice(ctx, r, func() ([]ContainerSummary, error) {
		return r.inner.ListContainers(ctx)
	})
}

// ListImages returns human-readable image references.
func (r *RetryExecutor) ListImages(ctx context.Context) ([]string, error) {
	return retryList(ctx, r, func() ([]string, error) {
		return r.inner.ListImages(ctx)
	})
}

// ListVolumes returns human-readable volume references.
func (r *RetryExecutor) ListVolumes(ctx context.Context) ([]string, error) {
	return retryList(ctx, r, func() ([]string, error) {
		return r.inner.ListVolumes(ctx)
	})
}

// ListNetworks returns human-readable network references.
func (r *RetryExecutor) ListNetworks(ctx context.Context) ([]string, error) {
	return retryList(ctx, r, func() ([]string, error) {
		return r.inner.ListNetworks(ctx)
	})
}

// InspectNetwork returns indented JSON for a network.
func (r *RetryExecutor) InspectNetwork(ctx context.Context, id string) (string, error) {
	return retry(ctx, r, func() (string, error) {
		return r.inner.InspectNetwork(ctx, id)
	})
}

// PruneImages removes dangling images.
func (r *RetryExecutor) PruneImages(ctx context.Context) (string, error) {
	return retry(ctx, r, func() (string, error) {
		return r.inner.PruneImages(ctx)
	})
}

// PruneVolumes removes unused volumes.
func (r *RetryExecutor) PruneVolumes(ctx context.Context) (string, error) {
	return retry(ctx, r, func() (string, error) {
		return r.inner.PruneVolumes(ctx)
	})
}

// PruneNetworks removes unused networks.
func (r *RetryExecutor) PruneNetworks(ctx context.Context) (string, error) {
	return retry(ctx, r, func() (string, error) {
		return r.inner.PruneNetworks(ctx)
	})
}

// Info returns Docker daemon information.
func (r *RetryExecutor) Info(ctx context.Context) (string, error) {
	return retry(ctx, r, func() (string, error) {
		return r.inner.Info(ctx)
	})
}

func retryList(ctx context.Context, r *RetryExecutor, fn func() ([]string, error)) ([]string, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err
		if attempt == maxRetries || !isRetryable(err) {
			break
		}

		if err := r.sleep(ctx, attempt); err != nil {
			return nil, err
		}
	}

	return nil, errors.Join(ErrMaxRetries, lastErr)
}

func retrySlice[T any](ctx context.Context, r *RetryExecutor, fn func() ([]T, error)) ([]T, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err
		if attempt == maxRetries || !isRetryable(err) {
			break
		}

		if err := r.sleep(ctx, attempt); err != nil {
			return nil, err
		}
	}

	return nil, errors.Join(ErrMaxRetries, lastErr)
}

func retry(ctx context.Context, r *RetryExecutor, fn func() (string, error)) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err
		if attempt == maxRetries || !isRetryable(err) {
			break
		}

		if err := r.sleep(ctx, attempt); err != nil {
			return "", err
		}
	}

	return "", errors.Join(ErrMaxRetries, lastErr)
}

var _ Executor = (*RetryExecutor)(nil)
