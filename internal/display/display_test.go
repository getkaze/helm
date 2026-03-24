package display

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/getkaze/helm/internal/config"
	"github.com/getkaze/helm/internal/session"
)

func init() {
	color.NoColor = true
}

func captureOutput(fn func()) string {
	var buf bytes.Buffer
	old := Out
	Out = &buf
	defer func() { Out = old }()
	fn()
	return buf.String()
}

func TestDashboard(t *testing.T) {
	s := &session.Session{
		Project:          session.Project{Name: "my-api", Type: "brownfield", State: "plan"},
		ExecutionProfile: "guided",
		CurrentAgent:     "research",
		Language:         "en-US",
		Agents: map[string]session.Agent{
			"survey":   {Status: "completed", Score: 100},
			"research": {Status: "in_progress"},
		},
	}
	cfg := &config.Config{Version: "v0.1.0"}

	output := captureOutput(func() { Dashboard(s, cfg) })

	if !strings.Contains(output, "my-api") {
		t.Error("dashboard should contain project name")
	}
	if !strings.Contains(output, "v0.1.0") {
		t.Error("dashboard should contain version")
	}
	if !strings.Contains(output, "[done]") {
		t.Error("dashboard should show [done] for completed agents")
	}
	if !strings.Contains(output, "[>>  ]") {
		t.Error("dashboard should show [>>  ] for current agent")
	}
	if !strings.Contains(output, "100%") {
		t.Error("dashboard should show score for completed agents")
	}
}

func TestShort(t *testing.T) {
	s := &session.Session{
		Project:      session.Project{Name: "my-api", Type: "brownfield", State: "plan"},
		CurrentAgent: "research",
		Agents: map[string]session.Agent{
			"research": {Status: "in_progress"},
		},
	}

	output := captureOutput(func() { Short(s) })
	expected := "my-api | plan | research (in_progress)"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q, got %q", expected, strings.TrimSpace(output))
	}
}

func TestJSON(t *testing.T) {
	s := &session.Session{
		Project:      session.Project{Name: "test", Type: "greenfield", State: "discover"},
		CurrentAgent: "scout",
		Language:     "en-US",
	}

	output := captureOutput(func() { JSON(s) })

	var parsed session.Session
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("JSON output is not valid JSON: %v", err)
	}
	if parsed.Project.Name != "test" {
		t.Errorf("expected name 'test', got %q", parsed.Project.Name)
	}
}

func TestInitSummary(t *testing.T) {
	output := captureOutput(func() { InitSummary("my-api", "brownfield", "pt-BR") })

	if !strings.Contains(output, "my-api") {
		t.Error("init summary should contain project name")
	}
	if !strings.Contains(output, "brownfield") {
		t.Error("init summary should contain project type")
	}
	if !strings.Contains(output, "pt-BR") {
		t.Error("init summary should contain language")
	}
	if !strings.Contains(output, "survey") {
		t.Error("init summary for brownfield should mention survey as next agent")
	}
}

func TestResumeContext(t *testing.T) {
	s := &session.Session{
		Project:      session.Project{Name: "my-api", Type: "brownfield", State: "plan"},
		CurrentAgent: "planning",
	}
	decisions := []string{"Chose Cobra", "Chose yaml.v3"}

	output := captureOutput(func() {
		ResumeContext(s, "research", 100, decisions, "Analyzed codebase")
	})

	if !strings.Contains(output, "research") {
		t.Error("should show last completed agent")
	}
	if !strings.Contains(output, "100%") {
		t.Error("should show score")
	}
	if !strings.Contains(output, "planning") {
		t.Error("should show current agent")
	}
	if !strings.Contains(output, "Chose Cobra") {
		t.Error("should show key decisions")
	}
	if !strings.Contains(output, "Analyzed codebase") {
		t.Error("should show summary context")
	}
	if !strings.Contains(output, "/helm") {
		t.Error("should show Claude Code instruction")
	}
}

func TestResumeContextEmpty(t *testing.T) {
	s := &session.Session{
		Project:      session.Project{Name: "test", Type: "greenfield", State: "discover"},
		CurrentAgent: "scout",
	}

	output := captureOutput(func() {
		ResumeContext(s, "", 0, nil, "")
	})

	if !strings.Contains(output, "scout") {
		t.Error("should show current agent even with empty context")
	}
}

func TestResumeFresh(t *testing.T) {
	s := &session.Session{
		Project:      session.Project{Name: "test", Type: "brownfield", State: "discover"},
		CurrentAgent: "survey",
	}

	output := captureOutput(func() { ResumeFresh(s) })

	if !strings.Contains(output, "No agents have run yet") {
		t.Error("should show fresh message")
	}
	if !strings.Contains(output, "survey") {
		t.Error("should show first agent")
	}
}

func TestSaveConfirmation(t *testing.T) {
	s := &session.Session{
		Project:      session.Project{Name: "test", Type: "brownfield", State: "build"},
		CurrentAgent: "build",
	}
	warnings := []string{"Handoff for scout not found."}

	output := captureOutput(func() {
		SaveConfirmation("2026-03-24T12-00-00", s, 5, 3, warnings)
	})

	if !strings.Contains(output, "2026-03-24T12-00-00") {
		t.Error("should show timestamp")
	}
	if !strings.Contains(output, "5 validated") {
		t.Error("should show handoff count")
	}
	if !strings.Contains(output, "3 validated") {
		t.Error("should show artifact count")
	}
	if !strings.Contains(output, "scout not found") {
		t.Error("should show warnings")
	}
	if !strings.Contains(output, "Safe to close") {
		t.Error("should show safe to close message")
	}
}

func TestSaveConfirmationNoWarnings(t *testing.T) {
	s := &session.Session{
		Project:      session.Project{Name: "test", Type: "brownfield", State: "build"},
		CurrentAgent: "build",
	}

	output := captureOutput(func() {
		SaveConfirmation("2026-03-24T12-00-00", s, 5, 3, nil)
	})

	if !strings.Contains(output, "Safe to close") {
		t.Error("should show safe to close")
	}
}

func TestWarning(t *testing.T) {
	output := captureOutput(func() { Warning("something went wrong") })
	if !strings.Contains(output, "something went wrong") {
		t.Error("warning should contain message")
	}
}

func TestDashboardNilConfig(t *testing.T) {
	s := &session.Session{
		Project:          session.Project{Name: "test", Type: "greenfield", State: "discover"},
		ExecutionProfile: "guided",
		CurrentAgent:     "scout",
		Language:         "en-US",
	}

	output := captureOutput(func() { Dashboard(s, nil) })
	if !strings.Contains(output, "test") {
		t.Error("dashboard should work with nil config")
	}
}

func TestDashboardWithTradeoffs(t *testing.T) {
	s := &session.Session{
		Project:          session.Project{Name: "test", Type: "brownfield", State: "plan"},
		ExecutionProfile: "guided",
		CurrentAgent:     "planning",
		Language:         "en-US",
		Tradeoffs: []session.Tradeoff{
			{Decision: "A vs B", Chosen: "A", Agent: "survey", Timestamp: "2026-03-24"},
		},
		Deviations: []session.Deviation{
			{Timestamp: "2026-03-24", Type: "scope_change", Reason: "added save"},
		},
	}

	output := captureOutput(func() { Dashboard(s, nil) })
	if !strings.Contains(output, "Tradeoffs: 1") {
		t.Error("should show tradeoff count")
	}
	if !strings.Contains(output, "Deviations: 1") {
		t.Error("should show deviation count")
	}
}

func TestError(t *testing.T) {
	// Error writes to os.Stderr, just verify it doesn't panic
	Error("test error message")
}

func TestInitSummaryGreenfield(t *testing.T) {
	output := captureOutput(func() { InitSummary("new-app", "greenfield", "en-US") })
	if !strings.Contains(output, "scout") {
		t.Error("greenfield should mention scout as first agent")
	}
}

func TestShortPendingAgent(t *testing.T) {
	s := &session.Session{
		Project:      session.Project{Name: "test", Type: "greenfield", State: "discover"},
		CurrentAgent: "scout",
	}
	output := captureOutput(func() { Short(s) })
	if !strings.Contains(output, "pending") {
		t.Error("unknown agent status should default to pending")
	}
}
