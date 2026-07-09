package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Title != "Ops Ronin" {
		t.Errorf("expected title Ops Ronin, got %s", cfg.Title)
	}
	if cfg.SocketPath != "/var/run/docker.sock" {
		t.Errorf("expected default socket, got %s", cfg.SocketPath)
	}
}

func TestLoadValidConfig(t *testing.T) {
	yaml := `
title: Test App
containers:
  - name: web
    label: Web Server
    commands:
      - name: logs
        label: View Logs
        action: logs
        args: ["--tail", "100"]
      - name: shell
        label: Open Shell
        action: exec
        command: sh
      - name: restart
        label: Restart
        action: restart
`
	cfg, err := LoadBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Title != "Test App" {
		t.Errorf("expected title Test App, got %s", cfg.Title)
	}
	if len(cfg.Containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(cfg.Containers))
	}
	if len(cfg.Containers[0].Commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(cfg.Containers[0].Commands))
	}
}

func TestLoadNewActions(t *testing.T) {
	yaml := `
containers:
  - name: web
    commands:
      - name: live-logs
        label: Live Logs
        action: logs-follow
      - name: stats
        label: Stats
        action: stats
      - name: monitor
        label: Monitor
        action: monitor
      - name: images
        label: Images
        action: images
      - name: volumes
        label: Volumes
        action: volumes
      - name: networks
        label: Networks
        action: networks
`
	cfg, err := LoadBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.Containers[0].Commands) != 6 {
		t.Fatalf("expected 6 commands, got %d", len(cfg.Containers[0].Commands))
	}
}

func TestLoadMissingContainers(t *testing.T) {
	yaml := `title: Empty`
	_, err := LoadBytes([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for missing containers")
	}
	if !strings.Contains(err.Error(), ErrNoContainers.Error()) {
		t.Errorf("expected no containers error, got %v", err)
	}
}

func TestLoadInvalidContainerName(t *testing.T) {
	yaml := `
containers:
  - name: "web;rm -rf"
    commands:
      - name: logs
        label: Logs
        action: logs
`
	_, err := LoadBytes([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for invalid container name")
	}
}

func TestLoadDuplicateContainer(t *testing.T) {
	yaml := `
containers:
  - name: web
    commands:
      - name: logs
        label: Logs
        action: logs
  - name: web
    commands:
      - name: logs
        label: Logs
        action: logs
`
	_, err := LoadBytes([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for duplicate container")
	}
}

func TestLoadDuplicateCommand(t *testing.T) {
	yaml := `
containers:
  - name: web
    commands:
      - name: logs
        label: Logs
        action: logs
      - name: logs
        label: Logs2
        action: status
`
	_, err := LoadBytes([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for duplicate command")
	}
}

func TestLoadInvalidSocketPath(t *testing.T) {
	yaml := `
socket: "../var/run/docker.sock"
containers:
  - name: web
    commands:
      - name: logs
        label: Logs
        action: logs
`
	_, err := LoadBytes([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for invalid socket path")
	}
}

func TestLoadMissingExecCommand(t *testing.T) {
	yaml := `
containers:
  - name: web
    commands:
      - name: shell
        label: Shell
        action: exec
`
	_, err := LoadBytes([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for missing exec command")
	}
}

func TestLoadInvalidAction(t *testing.T) {
	yaml := `
containers:
  - name: web
    commands:
      - name: hack
        label: Hack
        action: hack
`
	_, err := LoadBytes([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for invalid action")
	}
}

func TestCommandIsDangerous(t *testing.T) {
	tests := []struct {
		name     string
		cmd      Command
		expected bool
	}{
		{
			name:     "prune action",
			cmd:      Command{Name: "clean", Label: "Clean", Action: ActionPrune},
			expected: true,
		},
		{
			name:     "remove action",
			cmd:      Command{Name: "rm", Label: "Remove", Action: ActionRemove},
			expected: true,
		},
		{
			name:     "stop action",
			cmd:      Command{Name: "stop", Label: "Stop", Action: ActionStop},
			expected: true,
		},
		{
			name:     "dangerous keyword in label",
			cmd:      Command{Name: "dropdb", Label: "Drop Database", Action: ActionExec, Command: "echo"},
			expected: true,
		},
		{
			name:     "explicit dangerous flag",
			cmd:      Command{Name: "restart", Label: "Restart", Action: ActionRestart, Dangerous: true},
			expected: true,
		},
		{
			name:     "safe logs action",
			cmd:      Command{Name: "logs", Label: "Logs", Action: ActionLogs},
			expected: false,
		},
		{
			name:     "safe monitor action",
			cmd:      Command{Name: "monitor", Label: "Monitor", Action: ActionMonitor},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.IsDangerous(); got != tt.expected {
				t.Errorf("IsDangerous() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLoadSanitizesArgs(t *testing.T) {
	yaml := `
containers:
  - name: web
    commands:
      - name: logs
        label: Logs
        action: logs
        args: ["--tail", "100"]
`
	cfg, err := LoadBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Containers[0].Commands[0].Args[0] != "--tail" {
		t.Errorf("expected arg to be sanitized")
	}
}

func TestLoadRejectsInjectedArgs(t *testing.T) {
	yaml := `
containers:
  - name: web
    commands:
      - name: logs
        label: Logs
        action: logs
        args: ["--tail", "100;rm -rf /"]
`
	_, err := LoadBytes([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for injected args")
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "menu.yaml")

	yaml := `
title: File Test
containers:
  - name: app
    commands:
      - name: logs
        label: Logs
        action: logs
`
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Title != "File Test" {
		t.Errorf("expected title File Test, got %s", cfg.Title)
	}
}

func TestLoadFromMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/menu.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestValidateSetsDefaults(t *testing.T) {
	cfg := &Config{
		Containers: []Container{
			{
				Name: "web",
				Commands: []Command{
					{Name: "logs", Label: "Logs", Action: ActionLogs},
				},
			},
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.SocketPath != "/var/run/docker.sock" {
		t.Errorf("expected default socket, got %s", cfg.SocketPath)
	}

	if cfg.Containers[0].Label != "web" {
		t.Errorf("expected label to default to name, got %s", cfg.Containers[0].Label)
	}
}
