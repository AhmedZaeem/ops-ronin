package internal

import (
	"strings"
	"testing"
	"time"
)

func TestExecutionResult(t *testing.T) {
	result := ExecutionResult{
		Stdout:   "test output",
		Stderr:   "test error",
		ExitCode: 1,
	}

	if result.Stdout != "test output" {
		t.Errorf("Expected stdout 'test output', got '%s'", result.Stdout)
	}
	if result.Stderr != "test error" {
		t.Errorf("Expected stderr 'test error', got '%s'", result.Stderr)
	}
	if result.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", result.ExitCode)
	}
}

func TestExecuteCommand_EmptyContainerName(t *testing.T) {
	_, err := ExecuteCommand("", "echo test")
	if err == nil {
		t.Error("Expected error for empty container name")
	} else if !strings.Contains(err.Error(), "container name required") {
		t.Errorf("Expected 'container name required' error, got '%s'", err.Error())
	}
}

func TestExecuteCommand_EmptyCommand(t *testing.T) {
	_, err := ExecuteCommand("test-container", "")
	if err == nil {
		t.Error("Expected error for empty command")
	} else if !strings.Contains(err.Error(), "command required") {
		t.Errorf("Expected 'command required' error, got '%s'", err.Error())
	}
}

func TestExecuteCommandDetailed_EmptyContainerName(t *testing.T) {
	_, err := ExecuteCommandDetailed("", "echo test", 5*time.Second)
	if err == nil {
		t.Error("Expected error for empty container name")
	} else if !strings.Contains(err.Error(), "container name required") {
		t.Errorf("Expected 'container name required' error, got '%s'", err.Error())
	}
}

func TestExecuteCommandDetailed_EmptyCommand(t *testing.T) {
	_, err := ExecuteCommandDetailed("test-container", "", 5*time.Second)
	if err == nil {
		t.Error("Expected error for empty command")
	} else if !strings.Contains(err.Error(), "command required") {
		t.Errorf("Expected 'command required' error, got '%s'", err.Error())
	}
}

func TestExecuteCommandInteractive_EmptyContainerName(t *testing.T) {
	err := ExecuteCommandInteractive("", "echo test")
	if err == nil {
		t.Error("Expected error for empty container name")
	} else if !strings.Contains(err.Error(), "container name required") {
		t.Errorf("Expected 'container name required' error, got '%s'", err.Error())
	}
}

func TestExecuteCommandInteractive_EmptyCommand(t *testing.T) {
	err := ExecuteCommandInteractive("test-container", "")
	if err == nil {
		t.Error("Expected error for empty command")
	} else if !strings.Contains(err.Error(), "command required") {
		t.Errorf("Expected 'command required' error, got '%s'", err.Error())
	}
}
