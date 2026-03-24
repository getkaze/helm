package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/getkaze/helm/internal/display"
	"github.com/getkaze/helm/internal/session"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(oldDir) })
	return dir
}

func captureCmd(fn func()) string {
	var buf bytes.Buffer
	old := display.Out
	display.Out = &buf
	defer func() { display.Out = old }()
	fn()
	return buf.String()
}

func TestInitCreatesFiles(t *testing.T) {
	setupTestDir(t)

	initName = "test-project"
	initType = "greenfield"
	initLang = "en-US"
	initForce = false
	defer func() { initName = ""; initType = ""; initLang = "" }()

	err := runInit(nil, nil)
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Check directories created
	for _, dir := range []string{".helm", ".helm/handoffs", ".helm/artifacts", ".helm/checkpoints"} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("directory %s was not created", dir)
		}
	}

	// Check session.yaml
	s, err := session.Load(session.SessionFile)
	if err != nil {
		t.Fatalf("session not created: %v", err)
	}
	if s.Project.Name != "test-project" {
		t.Errorf("expected project name 'test-project', got %q", s.Project.Name)
	}
	if s.CurrentAgent != "scout" {
		t.Errorf("greenfield should start with scout, got %q", s.CurrentAgent)
	}

	// Check helm.yaml
	if _, err := os.Stat("helm.yaml"); os.IsNotExist(err) {
		t.Fatal("helm.yaml was not created")
	}
}

func TestInitBrownfieldRouting(t *testing.T) {
	setupTestDir(t)

	initName = "my-api"
	initType = "brownfield"
	initLang = "pt-BR"
	initForce = false
	defer func() { initName = ""; initType = ""; initLang = "" }()

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	s, _ := session.Load(session.SessionFile)
	if s.CurrentAgent != "survey" {
		t.Errorf("brownfield should start with survey, got %q", s.CurrentAgent)
	}
	if s.Language != "pt-BR" {
		t.Errorf("expected language pt-BR, got %q", s.Language)
	}
}

func TestInitExistingWithoutForce(t *testing.T) {
	dir := setupTestDir(t)
	os.MkdirAll(filepath.Join(dir, ".helm"), 0o755)

	initName = "test"
	initType = "greenfield"
	initLang = "en-US"
	initForce = false
	defer func() { initName = ""; initType = ""; initLang = "" }()

	// Redirect stdin to pipe so isInteractive() returns false
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Close()
	defer func() { os.Stdin = oldStdin; r.Close() }()

	err := runInit(nil, nil)
	if err == nil {
		t.Fatal("expected error when .helm/ exists and no --force in non-interactive mode")
	}
}

func TestInitExistingWithForce(t *testing.T) {
	dir := setupTestDir(t)
	os.MkdirAll(filepath.Join(dir, ".helm"), 0o755)

	initName = "test"
	initType = "greenfield"
	initLang = "en-US"
	initForce = true
	defer func() { initName = ""; initType = ""; initLang = ""; initForce = false }()

	err := runInit(nil, nil)
	if err != nil {
		t.Fatalf("init --force should succeed: %v", err)
	}
}

func TestInitInvalidType(t *testing.T) {
	setupTestDir(t)

	initName = "test"
	initType = "invalid"
	initLang = "en-US"
	defer func() { initName = ""; initType = ""; initLang = "" }()

	err := runInit(nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid project type")
	}
}

func TestStatusNoSession(t *testing.T) {
	setupTestDir(t)

	err := runStatus(nil, nil)
	if err == nil {
		t.Fatal("expected error when no session exists")
	}
}

func TestStatusWithSession(t *testing.T) {
	setupTestDir(t)

	// Create session
	os.MkdirAll(".helm", 0o755)
	s := &session.Session{
		Project:          session.Project{Name: "test", Type: "brownfield", State: "discover"},
		ExecutionProfile: "guided",
		CurrentAgent:     "survey",
		Language:         "en-US",
	}
	session.Save(session.SessionFile, s)

	statusJSON = false
	statusShort = false

	output := captureCmd(func() { runStatus(nil, nil) })
	if output == "" {
		t.Error("expected dashboard output")
	}
}

func TestStatusShort(t *testing.T) {
	setupTestDir(t)

	os.MkdirAll(".helm", 0o755)
	s := &session.Session{
		Project:      session.Project{Name: "test", Type: "greenfield", State: "discover"},
		CurrentAgent: "scout",
		Language:     "en-US",
	}
	session.Save(session.SessionFile, s)

	statusJSON = false
	statusShort = true
	defer func() { statusShort = false }()

	output := captureCmd(func() { runStatus(nil, nil) })
	if output == "" {
		t.Error("expected short output")
	}
}

func TestStatusJSON(t *testing.T) {
	setupTestDir(t)

	os.MkdirAll(".helm", 0o755)
	s := &session.Session{
		Project:      session.Project{Name: "test", Type: "greenfield", State: "discover"},
		CurrentAgent: "scout",
		Language:     "en-US",
	}
	session.Save(session.SessionFile, s)

	statusJSON = true
	statusShort = false
	defer func() { statusJSON = false }()

	output := captureCmd(func() { runStatus(nil, nil) })
	if output == "" {
		t.Error("expected JSON output")
	}
}

func TestVersionOutput(t *testing.T) {
	Version = "v0.1.0"
	defer func() { Version = "dev" }()

	output := captureCmd(func() {
		versionCmd.Run(versionCmd, nil)
	})
	if output != "helm v0.1.0\n" {
		t.Errorf("expected 'helm v0.1.0', got %q", output)
	}
}

func TestResumeNoSession(t *testing.T) {
	setupTestDir(t)

	resumeJSON = false
	err := runResume(nil, nil)
	if err == nil {
		t.Fatal("expected error when no session exists")
	}
}

func TestResumeFreshSession(t *testing.T) {
	setupTestDir(t)

	os.MkdirAll(".helm", 0o755)
	s := &session.Session{
		Project:      session.Project{Name: "test", Type: "greenfield", State: "discover"},
		CurrentAgent: "scout",
		Language:     "en-US",
	}
	session.Save(session.SessionFile, s)

	resumeJSON = false
	output := captureCmd(func() { runResume(nil, nil) })
	if output == "" {
		t.Error("expected fresh session output")
	}
}

func TestResumeWithHandoff(t *testing.T) {
	setupTestDir(t)

	os.MkdirAll(".helm/handoffs", 0o755)
	s := &session.Session{
		Project:      session.Project{Name: "test", Type: "brownfield", State: "plan"},
		CurrentAgent: "planning",
		Language:     "en-US",
		Agents: map[string]session.Agent{
			"survey": {Status: "completed", Score: 100},
		},
	}
	session.Save(session.SessionFile, s)

	handoffContent := `# Handoff: Survey → Research
## Summary
- **Agent**: survey
- **Status**: completed
- **Score**: 100%

### Mission Completed
Analyzed the codebase.

### Next Agent
→ **research**
`
	os.WriteFile(".helm/handoffs/survey.md", []byte(handoffContent), 0o644)

	resumeJSON = false
	output := captureCmd(func() { runResume(nil, nil) })
	if output == "" {
		t.Error("expected resume context output")
	}
}

func TestSaveNoSession(t *testing.T) {
	setupTestDir(t)

	saveForce = false
	err := runSave(nil, nil)
	if err == nil {
		t.Fatal("expected error when no session exists")
	}
}

func TestSaveCreatesCheckpoint(t *testing.T) {
	setupTestDir(t)

	os.MkdirAll(".helm/handoffs", 0o755)
	os.MkdirAll(".helm/artifacts/survey", 0o755)
	os.MkdirAll(".helm/checkpoints", 0o755)
	s := &session.Session{
		Project:          session.Project{Name: "test", Type: "brownfield", State: "plan"},
		ExecutionProfile: "guided",
		CurrentAgent:     "research",
		Language:         "en-US",
		Workflow:         "standard",
		Agents: map[string]session.Agent{
			"survey": {Status: "completed", Score: 100},
		},
	}
	session.Save(session.SessionFile, s)
	os.WriteFile(".helm/handoffs/survey.md", []byte("handoff"), 0o644)

	saveForce = true
	saveMessage = "test checkpoint"
	defer func() { saveForce = false; saveMessage = "" }()

	output := captureCmd(func() {
		err := runSave(nil, nil)
		if err != nil {
			t.Fatalf("save failed: %v", err)
		}
	})

	if output == "" {
		t.Error("expected save confirmation output")
	}

	// Check checkpoint was created
	entries, _ := os.ReadDir(".helm/checkpoints")
	if len(entries) == 0 {
		t.Fatal("no checkpoint file created")
	}
}

func TestErrNoSessionNotExist(t *testing.T) {
	err := errNoSession(os.ErrNotExist)
	if err == nil || err.Error() != "No Helm session found. Run 'helm init' first." {
		t.Errorf("expected friendly message, got: %v", err)
	}
}

func TestErrNoSessionOther(t *testing.T) {
	origErr := os.ErrPermission
	err := errNoSession(origErr)
	if err != origErr {
		t.Errorf("expected original error, got: %v", err)
	}
}

func TestDetectProjectType(t *testing.T) {
	dir := setupTestDir(t)

	// Empty dir = greenfield
	if detectProjectType() != "greenfield" {
		t.Error("empty dir should be greenfield")
	}

	// Create a .go file = brownfield
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o644)
	if detectProjectType() != "brownfield" {
		t.Error("dir with .go files should be brownfield")
	}
}

func TestDetectLanguage(t *testing.T) {
	// Default should be en-US
	os.Unsetenv("LANG")
	os.Unsetenv("LC_ALL")
	os.Unsetenv("LANGUAGE")
	if detectLanguage() != "en-US" {
		t.Error("default language should be en-US")
	}

	os.Setenv("LANG", "pt_BR.UTF-8")
	defer os.Unsetenv("LANG")
	if detectLanguage() != "pt-BR" {
		t.Error("pt in LANG should detect pt-BR")
	}
}

func TestCheckGitignore(t *testing.T) {
	setupTestDir(t)

	// No .gitignore — should not panic
	checkGitignore()

	// .gitignore without .helm/ — should warn
	os.WriteFile(".gitignore", []byte("node_modules/\n"), 0o644)
	output := captureCmd(func() { checkGitignore() })
	if output == "" {
		t.Error("should warn about missing .helm/ in .gitignore")
	}

	// .gitignore with .helm/ — should not warn
	os.WriteFile(".gitignore", []byte(".helm/\n"), 0o644)
	output = captureCmd(func() { checkGitignore() })
	if output != "" {
		t.Error("should not warn when .helm/ is in .gitignore")
	}
}
