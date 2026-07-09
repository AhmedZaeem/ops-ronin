package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/docker/docker/client"

	"github.com/AhmedZaeem/ops-ronin/internal/config"
	"github.com/AhmedZaeem/ops-ronin/internal/docker"
	"github.com/AhmedZaeem/ops-ronin/internal/tui"
)

const version = "v2.1.0"
const fallbackConfigPath = "menu.yaml"

func resolveDefaultConfig() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return fallbackConfigPath
	}

	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "menu.yaml")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return fallbackConfigPath
}

func main() {
	configPath := flag.String("config", resolveDefaultConfig(), "path to menu.yaml")
	showVersion := flag.Bool("version", false, "print version")
	flag.Parse()

	if *showVersion {
		fmt.Println("ops-ronin", version)
		os.Exit(0)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	cli, err := docker.NewClient(cfg.SocketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create docker client: %v\n", err)
		os.Exit(1)
	}
	defer cli.Close()

	if err := pingDocker(cli); err != nil {
		fmt.Fprintf(os.Stderr, "Docker daemon unreachable: %v\n", err)
		os.Exit(1)
	}

	base := docker.NewSDKExecutor(cli)
	executor := docker.NewRetryExecutor(base)

	model := tui.NewModel(cfg, executor)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}

func pingDocker(cli *client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := cli.Ping(ctx)
	return err
}
