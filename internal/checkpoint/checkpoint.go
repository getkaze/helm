package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/getkaze/helm/internal/session"
	"gopkg.in/yaml.v3"
)

type Checkpoint struct {
	Session   session.Session `yaml:"session"`
	Message   string          `yaml:"message,omitempty"`
	CreatedAt string          `yaml:"created_at"`
}

// Create writes a checkpoint file and returns the timestamp used.
func Create(dir string, s *session.Session, message string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create checkpoints directory: %w", err)
	}

	ts := time.Now().UTC().Format("2006-01-02T15-04-05")
	cp := Checkpoint{
		Session:   *s,
		Message:   message,
		CreatedAt: ts,
	}

	data, err := yaml.Marshal(&cp)
	if err != nil {
		return "", fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	path := filepath.Join(dir, ts+".yaml")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("failed to write checkpoint: %w", err)
	}

	return ts, nil
}

// List returns checkpoint filenames sorted by name (oldest first).
func List(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read checkpoints directory: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
			files = append(files, e.Name())
		}
	}

	sort.Strings(files)
	return files, nil
}

// Rotate deletes the oldest checkpoints if count exceeds max.
func Rotate(dir string, max int) error {
	files, err := List(dir)
	if err != nil {
		return err
	}

	for len(files) > max {
		path := filepath.Join(dir, files[0])
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove old checkpoint %s: %w", files[0], err)
		}
		files = files[1:]
	}

	return nil
}

// LatestTimestamp returns the timestamp of the most recent checkpoint, or empty string.
func LatestTimestamp(dir string) (string, time.Time, error) {
	files, err := List(dir)
	if err != nil || len(files) == 0 {
		return "", time.Time{}, err
	}

	latest := files[len(files)-1]
	name := strings.TrimSuffix(latest, ".yaml")
	t, err := time.Parse("2006-01-02T15-04-05", name)
	if err != nil {
		return name, time.Time{}, nil
	}
	return name, t, nil
}
