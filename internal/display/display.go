package display

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/getkaze/helm/internal/config"
	"github.com/getkaze/helm/internal/session"
)

var Out io.Writer = os.Stdout

var (
	green  = color.New(color.FgGreen)
	yellow = color.New(color.FgYellow)
	dim    = color.New(color.Faint)
	bold   = color.New(color.Bold)
	cyan   = color.New(color.FgCyan)
)

func Dashboard(s *session.Session, cfg *config.Config) {
	w := Out

	// Header
	version := ""
	if cfg != nil && cfg.Version != "" {
		version = " v" + cfg.Version
	}
	fmt.Fprintf(w, "\n")
	bold.Fprintf(w, "  Helm%s\n", version)
	fmt.Fprintf(w, "\n")

	// Project info
	fmt.Fprintf(w, "  Project:  %s\n", s.Project.Name)
	fmt.Fprintf(w, "  Type:     %s\n", s.Project.Type)
	fmt.Fprintf(w, "  Phase:    %s\n", s.Project.State)
	fmt.Fprintf(w, "  Profile:  %s\n", s.ExecutionProfile)
	fmt.Fprintf(w, "  Language: %s\n", s.Language)
	fmt.Fprintf(w, "\n")

	// Pipeline
	bold.Fprintf(w, "  Pipeline:\n")
	pipeline := session.Pipeline(s.Project.Type)
	for _, agent := range pipeline {
		a, exists := s.Agents[agent]
		if !exists {
			a = session.Agent{Status: "pending"}
		}

		indicator, scoreStr := formatAgent(agent, a, s.CurrentAgent)
		fmt.Fprintf(w, "    %s  %-12s %s\n", indicator, agent, scoreStr)
	}
	fmt.Fprintf(w, "\n")

	// Tradeoffs & deviations
	if len(s.Tradeoffs) > 0 || len(s.Deviations) > 0 {
		if len(s.Tradeoffs) > 0 {
			dim.Fprintf(w, "  Tradeoffs: %d\n", len(s.Tradeoffs))
		}
		if len(s.Deviations) > 0 {
			dim.Fprintf(w, "  Deviations: %d\n", len(s.Deviations))
		}
		fmt.Fprintf(w, "\n")
	}
}

func formatAgent(name string, a session.Agent, currentAgent string) (string, string) {
	switch {
	case a.Status == "completed":
		indicator := green.Sprint("[done]")
		scoreStr := green.Sprintf("%d%%", a.Score)
		return indicator, scoreStr
	case name == currentAgent || a.Status == "in_progress":
		indicator := yellow.Sprint("[>>  ]")
		scoreStr := yellow.Sprint("in progress")
		return indicator, scoreStr
	default:
		indicator := dim.Sprint("[    ]")
		scoreStr := dim.Sprint("pending")
		return indicator, scoreStr
	}
}

func Short(s *session.Session) {
	agentStatus := "pending"
	if a, ok := s.Agents[s.CurrentAgent]; ok {
		agentStatus = a.Status
	}
	fmt.Fprintf(Out, "%s | %s | %s (%s)\n", s.Project.Name, s.Project.State, s.CurrentAgent, agentStatus)
}

func InitSummary(name, projectType, lang string) {
	w := Out
	fmt.Fprintf(w, "\n")
	green.Fprintf(w, "  Helm initialized!\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "  Project:  %s\n", name)
	fmt.Fprintf(w, "  Type:     %s\n", projectType)
	fmt.Fprintf(w, "  Language: %s\n", lang)
	fmt.Fprintf(w, "\n")

	firstAgent := "scout"
	if projectType == "brownfield" {
		firstAgent = "survey"
	}
	cyan.Fprintf(w, "  Next step: use /helm in Claude Code to start with %s.\n", firstAgent)
	fmt.Fprintf(w, "\n")
}

func Error(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
}

func Warning(msg string) {
	yellow.Fprintf(Out, "  Warning: %s\n", msg)
}

func Prompt(question string) string {
	fmt.Fprintf(Out, "  %s ", question)
	var answer string
	fmt.Scanln(&answer)
	return strings.TrimSpace(answer)
}

// ResumeContext displays the resumption context for continuing work.
func ResumeContext(s *session.Session, lastAgent string, lastScore int, keyDecisions []string, summary string) {
	w := Out
	fmt.Fprintf(w, "\n")
	bold.Fprintf(w, "  Helm — Resume\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "  Project:  %s\n", s.Project.Name)
	fmt.Fprintf(w, "  Phase:    %s\n", s.Project.State)
	fmt.Fprintf(w, "\n")

	if lastAgent != "" {
		green.Fprintf(w, "  Last completed: %s (%d%%)\n", lastAgent, lastScore)
	}
	yellow.Fprintf(w, "  Current agent:  %s\n", s.CurrentAgent)
	fmt.Fprintf(w, "\n")

	if summary != "" {
		bold.Fprintf(w, "  Context:\n")
		fmt.Fprintf(w, "  %s\n", summary)
		fmt.Fprintf(w, "\n")
	}

	if len(keyDecisions) > 0 {
		bold.Fprintf(w, "  Key decisions:\n")
		for _, d := range keyDecisions {
			fmt.Fprintf(w, "    - %s\n", d)
		}
		fmt.Fprintf(w, "\n")
	}

	cyan.Fprintf(w, "  To continue, use /helm in Claude Code.\n")
	fmt.Fprintf(w, "\n")
}

// ResumeFresh displays the message for a fresh session with no completed agents.
func ResumeFresh(s *session.Session) {
	w := Out
	fmt.Fprintf(w, "\n")
	bold.Fprintf(w, "  Helm — Resume\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "  No agents have run yet. First agent: %s.\n", s.CurrentAgent)
	cyan.Fprintf(w, "  To start, use /helm in Claude Code.\n")
	fmt.Fprintf(w, "\n")
}

// SaveConfirmation displays the result of a save operation.
func SaveConfirmation(timestamp string, s *session.Session, handoffCount, artifactCount int, warnings []string) {
	w := Out
	fmt.Fprintf(w, "\n")
	green.Fprintf(w, "  Checkpoint saved!\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "  Timestamp: %s\n", timestamp)
	fmt.Fprintf(w, "  Phase:     %s\n", s.Project.State)
	fmt.Fprintf(w, "  Agent:     %s\n", s.CurrentAgent)
	fmt.Fprintf(w, "  Handoffs:  %d validated\n", handoffCount)
	fmt.Fprintf(w, "  Artifacts: %d validated\n", artifactCount)
	fmt.Fprintf(w, "\n")

	for _, warn := range warnings {
		Warning(warn)
	}
	if len(warnings) > 0 {
		fmt.Fprintf(w, "\n")
	}

	cyan.Fprintf(w, "  Safe to close this session.\n")
	fmt.Fprintf(w, "\n")
}
