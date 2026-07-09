package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/AhmedZaeem/ops-ronin/internal/docker"
)

type ContainerHealth int

const (
	HealthMissing ContainerHealth = iota
	HealthStopped
	HealthRunning
	HealthOther
)

type healthEntry struct {
	ConfigName string
	State      ContainerHealth
	ActualName string
	Status     string
	ID         string
}

type healthReport struct {
	entries []healthEntry
}

func (h ContainerHealth) Icon() string {
	switch h {
	case HealthRunning:
		return "✓"
	case HealthStopped:
		return "⏸"
	case HealthMissing:
		return "✗"
	default:
		return "?"
	}
}

func (h ContainerHealth) Color() string {
	switch h {
	case HealthRunning:
		return "#9ECE6A"
	case HealthStopped:
		return "#E0AF68"
	case HealthMissing:
		return "#F7768E"
	default:
		return "#A9B1D6"
	}
}

func checkContainerHealth(ctx context.Context, executor docker.Executor, configNames []string) (*healthReport, error) {
	containers, err := executor.ListContainers(ctx)
	if err != nil {
		return nil, err
	}

	byName := make(map[string]docker.ContainerSummary)
	for _, c := range containers {
		byName[c.Name] = c
	}

	report := &healthReport{entries: make([]healthEntry, 0, len(configNames))}
	for _, name := range configNames {
		entry := healthEntry{ConfigName: name}
		if actual, ok := byName[name]; ok {
			entry.ActualName = actual.Name
			entry.Status = actual.Status
			entry.ID = actual.ID
			switch strings.ToLower(actual.State) {
			case "running":
				entry.State = HealthRunning
			case "exited", "dead", "paused":
				entry.State = HealthStopped
			default:
				entry.State = HealthOther
			}
		} else {
			entry.State = HealthMissing
		}
		report.entries = append(report.entries, entry)
	}

	return report, nil
}

func findSuggestions(available []docker.ContainerSummary, missing string) []docker.ContainerSummary {
	lowerMissing := strings.ToLower(missing)
	var suggestions []docker.ContainerSummary

	for _, c := range available {
		lowerName := strings.ToLower(c.Name)
		if strings.Contains(lowerName, lowerMissing) || strings.Contains(lowerMissing, lowerName) {
			suggestions = append(suggestions, c)
		}
	}

	return suggestions
}

func applyFix(ctx context.Context, executor docker.Executor, entry healthEntry) (string, error) {
	switch entry.State {
	case HealthStopped:
		if err := executor.ContainerStart(ctx, entry.ActualName); err != nil {
			return "", fmt.Errorf("start container: %w", err)
		}
		return fmt.Sprintf("Container '%s' started", entry.ActualName), nil
	default:
		return "", fmt.Errorf("no automatic fix available")
	}
}
