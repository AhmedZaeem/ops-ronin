package monitoring

import (
	"sync"

	"github.com/AhmedZaeem/ops-ronin/internal/docker"
)

// RingBuffer stores a fixed number of stats samples in a thread-safe circular
// buffer suitable for sparkline rendering.
type RingBuffer struct {
	mu       sync.RWMutex
	capacity int
	items    []docker.StatsSample
	head     int
	full     bool
}

// NewRingBuffer creates a ring buffer with the requested capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity < 1 {
		capacity = 1
	}
	return &RingBuffer{
		capacity: capacity,
		items:    make([]docker.StatsSample, 0, capacity),
	}
}

// Push adds a sample to the buffer, evicting the oldest sample when full.
func (r *RingBuffer) Push(s docker.StatsSample) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.items) < r.capacity {
		r.items = append(r.items, s)
		r.head = len(r.items)
		if r.head == r.capacity {
			r.full = true
			r.head = 0
		}
		return
	}

	r.items[r.head] = s
	r.head++
	if r.head >= r.capacity {
		r.head = 0
	}
	r.full = true
}

// Snapshot returns the samples in chronological order.
func (r *RingBuffer) Snapshot() []docker.StatsSample {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.full {
		out := make([]docker.StatsSample, len(r.items))
		copy(out, r.items)
		return out
	}

	out := make([]docker.StatsSample, r.capacity)
	for i := 0; i < r.capacity; i++ {
		idx := (r.head + i) % r.capacity
		out[i] = r.items[idx]
	}
	return out
}

// Len returns the number of stored samples.
func (r *RingBuffer) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.full {
		return r.capacity
	}
	return len(r.items)
}

// CPUValues extracts CPU usage percentages from the buffer in order.
func (r *RingBuffer) CPUValues() []float64 {
	return extractValues(r.Snapshot(), func(s docker.StatsSample) float64 { return s.CPUUsage })
}

// MemoryValues extracts memory usage percentages from the buffer in order.
func (r *RingBuffer) MemoryValues() []float64 {
	return extractValues(r.Snapshot(), func(s docker.StatsSample) float64 { return s.MemoryPct })
}

func extractValues(samples []docker.StatsSample, fn func(docker.StatsSample) float64) []float64 {
	out := make([]float64, len(samples))
	for i, s := range samples {
		out[i] = fn(s)
	}
	return out
}
