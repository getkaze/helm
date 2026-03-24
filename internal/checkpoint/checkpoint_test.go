package checkpoint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getkaze/helm/internal/session"
)

func testSession() *session.Session {
	return &session.Session{
		Project:          session.Project{Name: "test", Type: "brownfield", State: "build"},
		ExecutionProfile: "guided",
		CurrentAgent:     "build",
		Language:         "en-US",
	}
}

func TestCreate(t *testing.T) {
	dir := t.TempDir()
	s := testSession()

	ts, err := Create(dir, s, "before arch changes")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if ts == "" {
		t.Fatal("expected non-empty timestamp")
	}

	// Check file exists
	path := filepath.Join(dir, ts+".yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("checkpoint file was not created")
	}
}

func TestListSorted(t *testing.T) {
	dir := t.TempDir()

	// Create files with known names (sorted alphabetically = chronologically)
	os.WriteFile(filepath.Join(dir, "2026-03-24T10-00-00.yaml"), []byte("a"), 0o644)
	os.WriteFile(filepath.Join(dir, "2026-03-24T11-00-00.yaml"), []byte("b"), 0o644)
	os.WriteFile(filepath.Join(dir, "2026-03-24T09-00-00.yaml"), []byte("c"), 0o644)

	files, err := List(dir)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}
	// Oldest first
	if files[0] != "2026-03-24T09-00-00.yaml" {
		t.Errorf("expected oldest first, got %s", files[0])
	}
	if files[2] != "2026-03-24T11-00-00.yaml" {
		t.Errorf("expected newest last, got %s", files[2])
	}
}

func TestRotateAtLimit(t *testing.T) {
	dir := t.TempDir()

	for i := 0; i < 6; i++ {
		name := filepath.Join(dir, "2026-03-24T10-0"+string(rune('0'+i))+"-00.yaml")
		os.WriteFile(name, []byte("data"), 0o644)
	}

	if err := Rotate(dir, 5); err != nil {
		t.Fatalf("Rotate failed: %v", err)
	}

	files, _ := List(dir)
	if len(files) != 5 {
		t.Errorf("expected 5 files after rotation, got %d", len(files))
	}
	// Oldest should be removed
	for _, f := range files {
		if f == "2026-03-24T10-00-00.yaml" {
			t.Error("oldest checkpoint should have been deleted")
		}
	}
}

func TestRotateUnderLimit(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "2026-03-24T10-00-00.yaml"), []byte("data"), 0o644)
	os.WriteFile(filepath.Join(dir, "2026-03-24T11-00-00.yaml"), []byte("data"), 0o644)

	if err := Rotate(dir, 5); err != nil {
		t.Fatalf("Rotate failed: %v", err)
	}

	files, _ := List(dir)
	if len(files) != 2 {
		t.Errorf("expected 2 files (under limit), got %d", len(files))
	}
}

func TestListNonexistentDir(t *testing.T) {
	files, err := List("/nonexistent/dir")
	if err != nil {
		t.Fatalf("expected nil error for nonexistent dir, got: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected empty list, got %d", len(files))
	}
}

func TestLatestTimestamp(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "2026-03-24T10-00-00.yaml"), []byte("a"), 0o644)
	os.WriteFile(filepath.Join(dir, "2026-03-24T12-00-00.yaml"), []byte("b"), 0o644)

	name, ts, err := LatestTimestamp(dir)
	if err != nil {
		t.Fatalf("LatestTimestamp failed: %v", err)
	}
	if name != "2026-03-24T12-00-00" {
		t.Errorf("expected latest name '2026-03-24T12-00-00', got %q", name)
	}
	if ts.IsZero() {
		t.Error("expected non-zero time")
	}
}

func TestLatestTimestampEmpty(t *testing.T) {
	dir := t.TempDir()

	name, ts, err := LatestTimestamp(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "" {
		t.Errorf("expected empty name, got %q", name)
	}
	if !ts.IsZero() {
		t.Error("expected zero time for empty dir")
	}
}

func TestCreateWithMessage(t *testing.T) {
	dir := t.TempDir()
	s := testSession()

	ts, err := Create(dir, s, "test message")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ts+".yaml"))
	if !strings.Contains(string(data), "test message") {
		t.Error("checkpoint should contain message")
	}
}
