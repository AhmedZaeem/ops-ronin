package internal

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	validConfig := `project: "test-project"
theme: "ronin"

menu:
  - title: "Test Operations"
    items:
      - label: "Test Command"
        container: "test-container"
        command: "echo hello"
`

	tmpFile, err := os.CreateTemp("", "test-menu-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(validConfig); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if cfg.Project != "test-project" {
		t.Errorf("Expected project 'test-project', got: %s", cfg.Project)
	}

	if len(cfg.Menu) != 1 {
		t.Errorf("Expected 1 menu category, got: %d", len(cfg.Menu))
	}

	if cfg.Menu[0].Title != "Test Operations" {
		t.Errorf("Expected title 'Test Operations', got: %s", cfg.Menu[0].Title)
	}
}

func TestLoadConfigInvalid(t *testing.T) {
	invalidYAML := `invalid: yaml: content: [
`

	tmpFile, err := os.CreateTemp("", "test-invalid-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(invalidYAML); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	_, err = LoadConfig(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML, got none")
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	_, err := LoadConfig("non-existent-file.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file, got none")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		expectedErrors int
	}{
		{
			name: "Valid config",
			config: &Config{
				Project: "test",
				Menu: []Category{{
					Title: "Test",
					Items: []Task{{
						Label:     "Test Task",
						Container: "test-container",
						Command:   "echo hello",
					}},
				}},
			},
			expectedErrors: 0,
		},
		{
			name: "Missing project",
			config: &Config{
				Menu: []Category{{
					Title: "Test",
					Items: []Task{{
						Label:     "Test Task",
						Container: "test-container",
						Command:   "echo hello",
					}},
				}},
			},
			expectedErrors: 1,
		},
		{
			name:           "Empty config",
			config:         &Config{},
			expectedErrors: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateConfig(tt.config)
			if len(errors) != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d: %v", tt.expectedErrors, len(errors), errors)
			}
		})
	}
}

func TestCreateExampleMenu(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-example-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	os.Remove(tmpFile.Name())

	err = CreateExampleMenu(tmpFile.Name())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Errorf("Created example menu should be valid, got error: %v", err)
	}

	if cfg.Project != "my-project" {
		t.Errorf("Example project should be 'my-project', got: %s", cfg.Project)
	}

	os.Remove(tmpFile.Name())
}
