package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/getkaze/helm/internal/session"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Version     string `yaml:"version"`
	InstalledAt string `yaml:"installed_at,omitempty"`
	ProjectType string `yaml:"project_type"`
	Language    string `yaml:"language"`
}

var (
	validProjectTypes = []string{"greenfield", "brownfield"}
	validLanguages    = []string{"en-US", "pt-BR"}
)

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("config file is corrupted: %w", err)
	}

	return &c, nil
}

func Save(path string, c *Config) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func Validate(c *Config) error {
	if c.Version == "" {
		return fmt.Errorf("config missing required field: version")
	}
	if c.ProjectType == "" {
		return fmt.Errorf("config missing required field: project_type")
	}
	if !session.Contains(validProjectTypes, c.ProjectType) {
		return fmt.Errorf("invalid project_type %q: must be one of %v", c.ProjectType, validProjectTypes)
	}
	if c.Language == "" {
		return fmt.Errorf("config missing required field: language")
	}
	if !session.Contains(validLanguages, c.Language) {
		return fmt.Errorf("invalid language %q: must be one of %v", c.Language, validLanguages)
	}
	return nil
}
