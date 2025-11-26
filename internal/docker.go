package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ExecutionResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func ExecuteCommand(containerName string, command string) (string, error) {
	result, err := ExecuteCommandDetailed(containerName, command, 30*time.Second)
	if err != nil {
		return "", err
	}
	return result.Stdout, nil
}

func ExecuteCommandDetailed(containerName string, command string, timeout time.Duration) (*ExecutionResult, error) {
	if containerName == "" {
		return nil, errors.New("container name required")
	}
	if command == "" {
		return nil, errors.New("command required")
	}

	if err := verifyContainer(containerName); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "exec", containerName, "sh", "-c", command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &ExecutionResult{
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
		ExitCode: 0,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute command: %w", err)
		}
	}

	return result, nil
}

func ExecuteCommandInteractive(containerName string, command string) error {
	if containerName == "" {
		return errors.New("container name required")
	}
	if command == "" {
		return errors.New("command required")
	}

	if err := verifyContainer(containerName); err != nil {
		return err
	}

	cmd := exec.Command("docker", "exec", "-it", containerName, "sh", "-c", command)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	return cmd.Run()
}

func verifyContainer(containerName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "inspect", containerName, "--format={{.State.Running}}")
	output, err := cmd.Output()
	if err != nil {
		dockerCmd := exec.CommandContext(ctx, "docker", "version")
		if dockerErr := dockerCmd.Run(); dockerErr != nil {
			return fmt.Errorf("Docker is not running or not accessible. Please start Docker and try again")
		}
		return fmt.Errorf("Container '%s' not found. Available containers can be listed with 'docker ps -a'", containerName)
	}

	running := strings.TrimSpace(string(output))
	if running != "true" {
		return fmt.Errorf("Container '%s' exists but is not running. Start it with 'docker start %s'", containerName, containerName)
	}

	return nil

}

func IsDockerAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "version", "--format={{.Server.Version}}")
	err := cmd.Run()
	return err == nil
}

func ExecuteCommandWithEnv(containerName, command string, env map[string]string, timeout time.Duration) (*ExecutionResult, error) {
	if containerName == "" {
		return nil, errors.New("container name required")
	}
	if command == "" {
		return nil, errors.New("command required")
	}

	if err := verifyContainer(containerName); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	args := []string{"exec"}
	for key, value := range env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}
	args = append(args, containerName, "sh", "-c", command)

	cmd := exec.CommandContext(ctx, "docker", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &ExecutionResult{
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
		ExitCode: 0,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute command: %w", err)
		}
	}

	return result, nil
}

func ListRunningContainers() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "ps", "--format={{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	containers := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(containers) == 1 && containers[0] == "" {
		return []string{}, nil
	}
	return containers, nil
}

func CopyToContainer(containerName, srcPath, destPath string) error {
	if err := verifyContainer(containerName); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "cp", srcPath, fmt.Sprintf("%s:%s", containerName, destPath))
	return cmd.Run()
}

func CopyFromContainer(containerName, srcPath, destPath string) error {
	if err := verifyContainer(containerName); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "cp", fmt.Sprintf("%s:%s", containerName, srcPath), destPath)
	return cmd.Run()
}
