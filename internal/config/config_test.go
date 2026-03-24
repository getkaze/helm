package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "helm.yaml")
	content := `version: v0.1.0
installed_at: "2026-03-24T12:00:00Z"
project_type: brownfield
language: pt-BR
`
	os.WriteFile(path, []byte(content), 0o644)

	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if c.Version != "v0.1.0" {
		t.Errorf("expected version 'v0.1.0', got %q", c.Version)
	}
	if c.ProjectType != "brownfield" {
		t.Errorf("expected project_type 'brownfield', got %q", c.ProjectType)
	}
	if c.Language != "pt-BR" {
		t.Errorf("expected language 'pt-BR', got %q", c.Language)
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := Load("/nonexistent/helm.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "helm.yaml")

	c := &Config{
		Version:     "dev",
		InstalledAt: "2026-03-24T12:00:00Z",
		ProjectType: "greenfield",
		Language:    "en-US",
	}

	if err := Save(path, c); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load after save failed: %v", err)
	}
	if loaded.Version != "dev" {
		t.Errorf("expected version 'dev', got %q", loaded.Version)
	}
}

func TestValidateValid(t *testing.T) {
	c := &Config{Version: "v1", ProjectType: "brownfield", Language: "en-US"}
	if err := Validate(c); err != nil {
		t.Fatalf("expected valid config, got: %v", err)
	}
}

func TestValidateMissingVersion(t *testing.T) {
	c := &Config{ProjectType: "brownfield", Language: "en-US"}
	if err := Validate(c); err == nil {
		t.Fatal("expected error for missing version")
	}
}

func TestValidateBadProjectType(t *testing.T) {
	c := &Config{Version: "v1", ProjectType: "invalid", Language: "en-US"}
	if err := Validate(c); err == nil {
		t.Fatal("expected error for invalid project type")
	}
}

func TestSaveCreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "helm.yaml")
	c := &Config{Version: "v1", ProjectType: "greenfield", Language: "en-US"}
	if err := Save(path, c); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("file was not created")
	}
}
