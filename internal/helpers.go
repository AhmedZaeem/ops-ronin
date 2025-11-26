package internal

import (
	"fmt"
	"os"
)

func CreateExampleMenu(path string) error {
	example := `project: "my-project"
theme: "ronin"

menu:
  - title: "Database Operations"
    items:
      - label: "Check Database Status"
        container: "my-database"
        command: "pg_isready -h localhost -p 5432"

      - label: "Show Database Version"
        container: "my-database"
        command: "psql --version"

  - title: "Application Operations"
    items:
      - label: "Check App Health"
        container: "my-app"
        command: "curl -f http://localhost:8080/health || echo 'Health check failed'"

      - label: "View Logs"
        container: "my-app"
        command: "tail -n 20 /var/log/app.log"

  - title: "System Operations"
    items:
      - label: "Check Disk Usage"
        container: "my-app"
        command: "df -h"

      - label: "Check Memory Usage"
        container: "my-app"
        command: "free -m"
`

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("menu.yaml already exists. Delete it first or use a different path")
	}

	return os.WriteFile(path, []byte(example), 0644)
}

func ValidateConfig(cfg *Config) []string {
	var errors []string

	if cfg.Project == "" {
		errors = append(errors, "Project name is required")
	}

	if len(cfg.Menu) == 0 {
		errors = append(errors, "At least one menu category is required")
	}

	for i, category := range cfg.Menu {
		if category.Title == "" {
			errors = append(errors, fmt.Sprintf("Category %d: title is required", i+1))
		}

		if len(category.Items) == 0 {
			errors = append(errors, fmt.Sprintf("Category '%s': at least one item is required", category.Title))
		}

		for j, item := range category.Items {
			if item.Label == "" {
				errors = append(errors, fmt.Sprintf("Category '%s', Item %d: label is required", category.Title, j+1))
			}
			if item.Container == "" {
				errors = append(errors, fmt.Sprintf("Category '%s', Item '%s': container is required", category.Title, item.Label))
			}
			if item.Command == "" {
				errors = append(errors, fmt.Sprintf("Category '%s', Item '%s': command is required", category.Title, item.Label))
			}
		}
	}

	return errors
}
