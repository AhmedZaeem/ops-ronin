package monitoring

import (
	"testing"
	"time"
	"unicode/utf8"

	"github.com/AhmedZaeem/ops-ronin/internal/docker"
)

func TestRingBufferPushAndSnapshot(t *testing.T) {
	r := NewRingBuffer(3)

	r.Push(docker.StatsSample{CPUUsage: 1})
	r.Push(docker.StatsSample{CPUUsage: 2})

	if r.Len() != 2 {
		t.Fatalf("expected length 2, got %d", r.Len())
	}

	snap := r.Snapshot()
	if len(snap) != 2 || snap[0].CPUUsage != 1 || snap[1].CPUUsage != 2 {
		t.Fatalf("unexpected snapshot: %v", snap)
	}

	r.Push(docker.StatsSample{CPUUsage: 3})
	r.Push(docker.StatsSample{CPUUsage: 4})

	if r.Len() != 3 {
		t.Fatalf("expected length 3, got %d", r.Len())
	}

	snap = r.Snapshot()
	if snap[0].CPUUsage != 2 || snap[1].CPUUsage != 3 || snap[2].CPUUsage != 4 {
		t.Fatalf("expected eviction of oldest, got: %v", snap)
	}
}

func TestRingBufferCPUValues(t *testing.T) {
	r := NewRingBuffer(2)
	r.Push(docker.StatsSample{CPUUsage: 10, Timestamp: time.Now()})
	r.Push(docker.StatsSample{CPUUsage: 20, Timestamp: time.Now()})

	values := r.CPUValues()
	if len(values) != 2 || values[0] != 10 || values[1] != 20 {
		t.Fatalf("unexpected cpu values: %v", values)
	}
}

func TestSparkline(t *testing.T) {
	values := []float64{0, 25, 50, 75, 100}
	line := Sparkline(values)
	if utf8.RuneCountInString(line) != len(values) {
		t.Fatalf("expected sparkline length %d, got %d", len(values), utf8.RuneCountInString(line))
	}
}

func TestSparklineEmpty(t *testing.T) {
	if Sparkline(nil) != "" {
		t.Fatal("expected empty sparkline for empty input")
	}
}

func TestSparklineConstant(t *testing.T) {
	values := []float64{5, 5, 5}
	line := Sparkline(values)
	if utf8.RuneCountInString(line) != len(values) {
		t.Fatalf("expected constant sparkline length %d, got %d", len(values), utf8.RuneCountInString(line))
	}
}

func TestFormatBytes(t *testing.T) {
	if FormatBytes(512) != "512 B" {
		t.Errorf("unexpected byte format: %s", FormatBytes(512))
	}
	if FormatBytes(1024) != "1.0 KiB" {
		t.Errorf("unexpected KiB format: %s", FormatBytes(1024))
	}
}
