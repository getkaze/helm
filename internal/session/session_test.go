package session

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.yaml")
	content := `project:
  name: "test-project"
  type: brownfield
  state: discover
execution_profile: guided
current_agent: survey
language: en-US
workflow: standard
agents:
  survey:
    status: completed
    score: 95
    criteria_count: 7
    completed_at: "2026-03-24"
`
	os.WriteFile(path, []byte(content), 0o644)

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if s.Project.Name != "test-project" {
		t.Errorf("expected project name 'test-project', got %q", s.Project.Name)
	}
	if s.Project.Type != "brownfield" {
		t.Errorf("expected project type 'brownfield', got %q", s.Project.Type)
	}
	if s.CurrentAgent != "survey" {
		t.Errorf("expected current_agent 'survey', got %q", s.CurrentAgent)
	}
	if a, ok := s.Agents["survey"]; !ok {
		t.Error("expected survey agent in agents map")
	} else if a.Score != 95 {
		t.Errorf("expected survey score 95, got %d", a.Score)
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := Load("/nonexistent/session.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.yaml")
	os.WriteFile(path, []byte("{{invalid yaml"), 0o644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestSaveAndBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.yaml")

	// Write initial file
	os.WriteFile(path, []byte("old content"), 0o644)

	s := &Session{
		Project:          Project{Name: "test", Type: "greenfield", State: "discover"},
		ExecutionProfile: "guided",
		CurrentAgent:     "scout",
		Language:         "en-US",
	}

	if err := Save(path, s); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Check backup was created
	backupPath := path + ".backup"
	backup, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("backup file not created: %v", err)
	}
	if string(backup) != "old content" {
		t.Errorf("backup content mismatch: got %q", string(backup))
	}

	// Check saved file can be loaded back
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("failed to load saved file: %v", err)
	}
	if loaded.Project.Name != "test" {
		t.Errorf("expected name 'test', got %q", loaded.Project.Name)
	}
}

func TestValidateValid(t *testing.T) {
	s := &Session{
		Project:          Project{Name: "test", Type: "brownfield", State: "discover"},
		ExecutionProfile: "guided",
		CurrentAgent:     "survey",
		Language:         "pt-BR",
		Workflow:         "standard",
	}
	if err := Validate(s); err != nil {
		t.Fatalf("expected valid session, got error: %v", err)
	}
}

func TestValidateMissingName(t *testing.T) {
	s := &Session{
		Project:      Project{Name: "", Type: "brownfield", State: "discover"},
		CurrentAgent: "survey",
		Language:     "en-US",
	}
	err := Validate(s)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestValidateBadState(t *testing.T) {
	s := &Session{
		Project:      Project{Name: "test", Type: "brownfield", State: "invalid"},
		CurrentAgent: "survey",
		Language:     "en-US",
	}
	err := Validate(s)
	if err == nil {
		t.Fatal("expected error for invalid state")
	}
}

func TestValidateBadAgentStatus(t *testing.T) {
	s := &Session{
		Project:      Project{Name: "test", Type: "brownfield", State: "discover"},
		CurrentAgent: "survey",
		Language:     "en-US",
		Agents: map[string]Agent{
			"survey": {Status: "bad_status"},
		},
	}
	err := Validate(s)
	if err == nil {
		t.Fatal("expected error for bad agent status")
	}
}

func TestSaveNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "session.yaml")

	s := &Session{
		Project:      Project{Name: "test", Type: "greenfield", State: "discover"},
		CurrentAgent: "scout",
		Language:     "en-US",
	}

	if err := Save(path, s); err != nil {
		t.Fatalf("Save to new dir failed: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("file was not created")
	}
}

func TestValidateMissingType(t *testing.T) {
	s := &Session{
		Project:      Project{Name: "test", Type: "", State: "discover"},
		CurrentAgent: "survey",
		Language:     "en-US",
	}
	if err := Validate(s); err == nil {
		t.Fatal("expected error for missing type")
	}
}

func TestValidateBadType(t *testing.T) {
	s := &Session{
		Project:      Project{Name: "test", Type: "invalid", State: "discover"},
		CurrentAgent: "survey",
		Language:     "en-US",
	}
	if err := Validate(s); err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestValidateMissingAgent(t *testing.T) {
	s := &Session{
		Project:  Project{Name: "test", Type: "brownfield", State: "discover"},
		Language: "en-US",
	}
	if err := Validate(s); err == nil {
		t.Fatal("expected error for missing current_agent")
	}
}

func TestValidateBadAgent(t *testing.T) {
	s := &Session{
		Project:      Project{Name: "test", Type: "brownfield", State: "discover"},
		CurrentAgent: "nonexistent",
		Language:     "en-US",
	}
	if err := Validate(s); err == nil {
		t.Fatal("expected error for invalid agent")
	}
}

func TestValidateMissingLanguage(t *testing.T) {
	s := &Session{
		Project:      Project{Name: "test", Type: "brownfield", State: "discover"},
		CurrentAgent: "survey",
	}
	if err := Validate(s); err == nil {
		t.Fatal("expected error for missing language")
	}
}

func TestValidateBadLanguage(t *testing.T) {
	s := &Session{
		Project:      Project{Name: "test", Type: "brownfield", State: "discover"},
		CurrentAgent: "survey",
		Language:     "fr-FR",
	}
	if err := Validate(s); err == nil {
		t.Fatal("expected error for invalid language")
	}
}

func TestValidateBadProfile(t *testing.T) {
	s := &Session{
		Project:          Project{Name: "test", Type: "brownfield", State: "discover"},
		ExecutionProfile: "invalid",
		CurrentAgent:     "survey",
		Language:         "en-US",
	}
	if err := Validate(s); err == nil {
		t.Fatal("expected error for invalid execution_profile")
	}
}

func TestValidateBadWorkflow(t *testing.T) {
	s := &Session{
		Project:      Project{Name: "test", Type: "brownfield", State: "discover"},
		CurrentAgent: "survey",
		Language:     "en-US",
		Workflow:     "invalid",
	}
	if err := Validate(s); err == nil {
		t.Fatal("expected error for invalid workflow")
	}
}

func TestValidateBadAgentName(t *testing.T) {
	s := &Session{
		Project:      Project{Name: "test", Type: "brownfield", State: "discover"},
		CurrentAgent: "survey",
		Language:     "en-US",
		Agents: map[string]Agent{
			"nonexistent_agent": {Status: "pending"},
		},
	}
	if err := Validate(s); err == nil {
		t.Fatal("expected error for invalid agent name in map")
	}
}

func TestBackupMissingFile(t *testing.T) {
	err := Backup("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestContains(t *testing.T) {
	if !Contains([]string{"a", "b", "c"}, "b") {
		t.Error("should find 'b'")
	}
	if Contains([]string{"a", "b"}, "z") {
		t.Error("should not find 'z'")
	}
}

func TestPipeline(t *testing.T) {
	gf := Pipeline("greenfield")
	if gf[0] != "scout" {
		t.Errorf("greenfield should start with scout, got %s", gf[0])
	}
	bf := Pipeline("brownfield")
	if bf[0] != "survey" {
		t.Errorf("brownfield should start with survey, got %s", bf[0])
	}
	if len(gf) != 10 || len(bf) != 10 {
		t.Errorf("pipelines should have 10 agents each")
	}
}
