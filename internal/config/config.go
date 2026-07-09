package config

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/AhmedZaeem/ops-ronin/internal/sanitize"
)

// Action represents a supported built-in operation.
type Action string

const (
	ActionLogs           Action = "logs"
	ActionLogsFollow     Action = "logs-follow"
	ActionExec           Action = "exec"
	ActionRestart        Action = "restart"
	ActionStop           Action = "stop"
	ActionStart          Action = "start"
	ActionStatus         Action = "status"
	ActionPrune          Action = "prune"
	ActionRemove         Action = "remove"
	ActionStats          Action = "stats"
	ActionMonitor        Action = "monitor"
	ActionImages         Action = "images"
	ActionVolumes        Action = "volumes"
	ActionNetworks       Action = "networks"
	ActionNetworkInspect Action = "network-inspect"
)

var validActions = map[Action]bool{
	ActionLogs:           true,
	ActionLogsFollow:     true,
	ActionExec:           true,
	ActionRestart:        true,
	ActionStop:           true,
	ActionStart:          true,
	ActionStatus:         true,
	ActionPrune:          true,
	ActionRemove:         true,
	ActionStats:          true,
	ActionMonitor:        true,
	ActionImages:         true,
	ActionVolumes:        true,
	ActionNetworks:       true,
	ActionNetworkInspect: true,
}

// dangerousActions are actions that always require explicit confirmation.
var dangerousActions = map[Action]bool{
	ActionPrune:  true,
	ActionRemove: true,
	ActionStop:   true,
}

// dangerousKeywords are matched against command names, labels, and exec commands
// to surface confirmation dialogs for destructive operations.
var dangerousKeywords = []string{"rm", "remove", "drop", "delete", "prune", "destroy", "kill", "stop"}

var containerNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)
var commandNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

var (
	ErrEmptyConfig        = errors.New("configuration is empty")
	ErrNoContainers       = errors.New("at least one container is required")
	ErrInvalidContainer   = errors.New("invalid container name")
	ErrNoCommands         = errors.New("at least one command is required")
	ErrInvalidCommand     = errors.New("invalid command name")
	ErrInvalidAction      = errors.New("invalid action")
	ErrMissingExecCommand = errors.New("exec action requires a command")
	ErrInvalidSocketPath  = errors.New("invalid docker socket path")
	ErrEmptyLabel         = errors.New("label cannot be empty")
	ErrDuplicateCommand   = errors.New("duplicate command name")
	ErrDuplicateContainer = errors.New("duplicate container name")
	ErrDangerousChars     = errors.New("input contains dangerous characters")
)

// Command defines a single operation that can be executed against a container
// or the Docker host.
type Command struct {
	Name      string   `yaml:"name"`
	Label     string   `yaml:"label"`
	Action    Action   `yaml:"action"`
	Command   string   `yaml:"command,omitempty"`
	Args      []string `yaml:"args,omitempty"`
	Dangerous bool     `yaml:"dangerous,omitempty"`
}

// Container groups commands that target a specific Docker container.
type Container struct {
	Name     string    `yaml:"name"`
	Label    string    `yaml:"label,omitempty"`
	Commands []Command `yaml:"commands"`
}

// Config is the top-level YAML configuration.
type Config struct {
	Title      string      `yaml:"title,omitempty"`
	SocketPath string      `yaml:"socket,omitempty"`
	Containers []Container `yaml:"containers"`
}

// IsDangerous returns true if the command should trigger a confirmation dialog.
func (c *Command) IsDangerous() bool {
	if c.Dangerous {
		return true
	}
	if dangerousActions[c.Action] {
		return true
	}
	lower := strings.ToLower(c.Name + " " + c.Label + " " + c.Command)
	for _, kw := range dangerousKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// DefaultConfig returns a baseline configuration.
func DefaultConfig() *Config {
	return &Config{
		Title:      "Ops Ronin",
		SocketPath: "/var/run/docker.sock",
		Containers: []Container{},
	}
}

// Validate sanitizes and validates the configuration.
func (cfg *Config) Validate() error {
	if cfg == nil {
		return ErrEmptyConfig
	}

	cfg.SocketPath = strings.TrimSpace(cfg.SocketPath)
	if cfg.SocketPath == "" {
		cfg.SocketPath = "/var/run/docker.sock"
	}

	if !strings.HasPrefix(cfg.SocketPath, "/") {
		return fmt.Errorf("%w: %s", ErrInvalidSocketPath, cfg.SocketPath)
	}

	if strings.Contains(cfg.SocketPath, "..") {
		return fmt.Errorf("%w: path traversal detected", ErrInvalidSocketPath)
	}

	if len(cfg.Containers) == 0 {
		return ErrNoContainers
	}

	seenContainers := make(map[string]bool)
	for i := range cfg.Containers {
		container := &cfg.Containers[i]
		container.Name = strings.TrimSpace(container.Name)
		container.Label = strings.TrimSpace(container.Label)

		if container.Label == "" {
			container.Label = container.Name
		}

		if container.Name == "" {
			return fmt.Errorf("%w at index %d", ErrInvalidContainer, i)
		}

		if !containerNameRegex.MatchString(container.Name) {
			return fmt.Errorf("%w: %s", ErrInvalidContainer, container.Name)
		}

		clean, err := sanitize.Identifier(container.Name)
		if err != nil || clean != container.Name {
			return fmt.Errorf("%w: %s", ErrInvalidContainer, container.Name)
		}

		if seenContainers[container.Name] {
			return fmt.Errorf("%w: %s", ErrDuplicateContainer, container.Name)
		}
		seenContainers[container.Name] = true

		if len(container.Commands) == 0 {
			return fmt.Errorf("%w for container %s", ErrNoCommands, container.Name)
		}

		seenCommands := make(map[string]bool)
		for j := range container.Commands {
			cmd := &container.Commands[j]
			if err := validateCommand(cmd); err != nil {
				return fmt.Errorf("container %s command %d: %w", container.Name, j, err)
			}
			if seenCommands[cmd.Name] {
				return fmt.Errorf("%w: %s", ErrDuplicateCommand, cmd.Name)
			}
			seenCommands[cmd.Name] = true
		}
	}

	return nil
}

func validateCommand(cmd *Command) error {
	cmd.Name = strings.TrimSpace(cmd.Name)
	cmd.Label = strings.TrimSpace(cmd.Label)
	cmd.Command = strings.TrimSpace(cmd.Command)

	if cmd.Name == "" {
		return ErrInvalidCommand
	}

	if !commandNameRegex.MatchString(cmd.Name) {
		return fmt.Errorf("%w: %s", ErrInvalidCommand, cmd.Name)
	}

	clean, err := sanitize.Identifier(cmd.Name)
	if err != nil || clean != cmd.Name {
		return fmt.Errorf("%w: %s", ErrInvalidCommand, cmd.Name)
	}

	if cmd.Label == "" {
		return ErrEmptyLabel
	}

	if _, ok := validActions[cmd.Action]; !ok {
		return fmt.Errorf("%w: %s", ErrInvalidAction, cmd.Action)
	}

	if cmd.Action == ActionExec && cmd.Command == "" {
		return ErrMissingExecCommand
	}

	for k, arg := range cmd.Args {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			return fmt.Errorf("empty argument at index %d", k)
		}
		cleaned, err := sanitize.Argument(arg)
		if err != nil {
			return fmt.Errorf("%w in args[%d]", err, k)
		}
		cmd.Args[k] = cleaned
	}

	if cmd.Command != "" {
		cleaned, err := sanitize.Argument(cmd.Command)
		if err != nil {
			return fmt.Errorf("%w in command", err)
		}
		cmd.Command = cleaned
	}

	return nil
}
