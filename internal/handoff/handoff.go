package handoff

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type HandoffSummary struct {
	Agent        string
	Status       string
	Score        int
	Timestamp    string
	Summary      string
	KeyDecisions []string
	NextAgent    string
}

// List returns handoff filenames sorted by modification time (newest first).
func List(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read handoffs directory: %w", err)
	}

	type fileTime struct {
		name    string
		modTime int64
	}

	var files []fileTime
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileTime{name: e.Name(), modTime: info.ModTime().UnixNano()})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime > files[j].modTime
	})

	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.name
	}
	return names, nil
}

// ReadLatest returns the content of the most recently modified handoff file.
func ReadLatest(dir string) (string, string, error) {
	files, err := List(dir)
	if err != nil {
		return "", "", err
	}
	if len(files) == 0 {
		return "", "", fmt.Errorf("no handoff files found")
	}

	path := filepath.Join(dir, files[0])
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("failed to read handoff file: %w", err)
	}

	return string(data), files[0], nil
}

// ParseSummary does best-effort parsing of a handoff markdown file.
func ParseSummary(content string) HandoffSummary {
	var h HandoffSummary

	// Agent name from header: "# Handoff: Agent → Next"
	if m := regexp.MustCompile(`#\s+Handoff:\s+(\w+)`).FindStringSubmatch(content); len(m) > 1 {
		h.Agent = strings.ToLower(m[1])
	}

	// Status
	if m := regexp.MustCompile(`\*\*Status\*\*:\s*(\w+)`).FindStringSubmatch(content); len(m) > 1 {
		h.Status = m[1]
	}

	// Score
	if m := regexp.MustCompile(`\*\*Score\*\*:\s*(\d+)`).FindStringSubmatch(content); len(m) > 1 {
		h.Score, _ = strconv.Atoi(m[1])
	}

	// Timestamp
	if m := regexp.MustCompile(`\*\*Timestamp\*\*:\s*(.+)`).FindStringSubmatch(content); len(m) > 1 {
		h.Timestamp = strings.TrimSpace(m[1])
	}

	// Mission Completed / Summary section
	if m := regexp.MustCompile(`(?i)###\s+Mission Completed\n([\s\S]*?)(?:\n###|\n---|\z)`).FindStringSubmatch(content); len(m) > 1 {
		h.Summary = strings.TrimSpace(m[1])
	}

	// Key Decisions or Key Findings
	re := regexp.MustCompile(`(?i)###\s+Key (?:Decisions|Findings)\n([\s\S]*?)(?:\n###|\n---|\z)`)
	if m := re.FindStringSubmatch(content); len(m) > 1 {
		lines := strings.Split(strings.TrimSpace(m[1]), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "- ") {
				h.KeyDecisions = append(h.KeyDecisions, strings.TrimPrefix(line, "- "))
			}
		}
	}

	// Next Agent
	if m := regexp.MustCompile(`(?i)(?:###\s+)?Next Agent\n.*?\*\*(\w+)\*\*`).FindStringSubmatch(content); len(m) > 1 {
		h.NextAgent = strings.ToLower(m[1])
	} else if m := regexp.MustCompile(`→\s+\*\*(\w+)\*\*`).FindStringSubmatch(content); len(m) > 1 {
		h.NextAgent = strings.ToLower(m[1])
	}

	return h
}
