package handoff

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const sampleHandoff = `# Handoff: Review → Build

## Summary
- **Agent**: review
- **Status**: completed
- **Score**: 100% (7/7 criteria met)
- **Timestamp**: 2026-03-24
- **Verdict**: APPROVED

### Mission Completed
Validated full traceability across all 6 planning artifacts.

### Key Findings
- 5/5 Research problems traced
- All artifacts consistent
- Dependency DAG is valid

### Next Agent
→ **build**
`

const partialHandoff = `# Handoff: Survey → Research

## Summary
- **Agent**: survey
- **Status**: completed
- **Score**: 95%
- **Timestamp**: 2026-03-24

### Mission Completed
Analyzed the codebase structure.

### Next Agent
→ **research**
`

func TestListSortedByTime(t *testing.T) {
	dir := t.TempDir()

	// Create files with different times
	os.WriteFile(filepath.Join(dir, "survey.md"), []byte("old"), 0o644)
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(filepath.Join(dir, "research.md"), []byte("new"), 0o644)

	files, err := List(dir)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0] != "research.md" {
		t.Errorf("expected newest first (research.md), got %s", files[0])
	}
}

func TestReadLatest(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "survey.md"), []byte("old content"), 0o644)
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(filepath.Join(dir, "research.md"), []byte("new content"), 0o644)

	content, filename, err := ReadLatest(dir)
	if err != nil {
		t.Fatalf("ReadLatest failed: %v", err)
	}
	if filename != "research.md" {
		t.Errorf("expected research.md, got %s", filename)
	}
	if content != "new content" {
		t.Errorf("expected 'new content', got %q", content)
	}
}

func TestReadLatestEmpty(t *testing.T) {
	dir := t.TempDir()
	_, _, err := ReadLatest(dir)
	if err == nil {
		t.Fatal("expected error for empty directory")
	}
}

func TestParseSummaryFull(t *testing.T) {
	h := ParseSummary(sampleHandoff)

	if h.Agent != "review" {
		t.Errorf("expected agent 'review', got %q", h.Agent)
	}
	if h.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", h.Status)
	}
	if h.Score != 100 {
		t.Errorf("expected score 100, got %d", h.Score)
	}
	if h.Summary == "" {
		t.Error("expected non-empty summary")
	}
	if len(h.KeyDecisions) == 0 {
		t.Error("expected key decisions to be parsed")
	}
	if h.NextAgent != "build" {
		t.Errorf("expected next agent 'build', got %q", h.NextAgent)
	}
}

func TestParseSummaryPartial(t *testing.T) {
	h := ParseSummary(partialHandoff)

	if h.Agent != "survey" {
		t.Errorf("expected agent 'survey', got %q", h.Agent)
	}
	if h.Score != 95 {
		t.Errorf("expected score 95, got %d", h.Score)
	}
	// No Key Decisions section — should be empty, not error
	if len(h.KeyDecisions) != 0 {
		t.Errorf("expected empty key decisions, got %d", len(h.KeyDecisions))
	}
	if h.NextAgent != "research" {
		t.Errorf("expected next agent 'research', got %q", h.NextAgent)
	}
}

func TestParseSummaryEmpty(t *testing.T) {
	h := ParseSummary("")
	if h.Agent != "" {
		t.Errorf("expected empty agent, got %q", h.Agent)
	}
}
