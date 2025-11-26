package internal

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Project string     `yaml:"project"`
	Theme   string     `yaml:"theme"`
	Menu    []Category `yaml:"menu"`
}

type Category struct {
	Title string `yaml:"title"`
	Items []Task `yaml:"items"`
}

type Task struct {
	Label     string `yaml:"label"`
	Container string `yaml:"container"`
	Command   string `yaml:"command"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if errors := ValidateConfig(&cfg); len(errors) > 0 {
		return nil, fmt.Errorf("configuration validation failed:\n- %s", strings.Join(errors, "\n- "))
	}

	return &cfg, nil
}
func (c *Config) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}
func SaveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
