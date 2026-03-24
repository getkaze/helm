package cmd

import (
	"github.com/getkaze/helm/internal/display"
	"github.com/getkaze/helm/internal/handoff"
	"github.com/getkaze/helm/internal/session"
	"github.com/spf13/cobra"
)

var resumeJSON bool

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Show session resumption context",
	Long:  "Display where you left off — last completed agent, current agent, and handoff context.",
	RunE:  runResume,
}

func init() {
	resumeCmd.Flags().BoolVar(&resumeJSON, "json", false, "output as JSON")
	rootCmd.AddCommand(resumeCmd)
}

func runResume(cmd *cobra.Command, args []string) error {
	s, err := session.Load(session.SessionFile)
	if err != nil {
		return errNoSession(err)
	}

	if resumeJSON {
		return display.JSON(s)
	}

	// Check if any agents have completed
	hasCompleted := false
	for _, a := range s.Agents {
		if a.Status == "completed" {
			hasCompleted = true
			break
		}
	}

	if !hasCompleted {
		display.ResumeFresh(s)
		return nil
	}

	// Find latest handoff
	content, _, err := handoff.ReadLatest(session.HandoffsDir)
	if err != nil {
		display.Warning("Handoff for " + s.CurrentAgent + " not found. Context may be incomplete.")
		display.ResumeContext(s, "", 0, nil, "")
		return nil
	}

	h := handoff.ParseSummary(content)

	display.ResumeContext(s, h.Agent, h.Score, h.KeyDecisions, h.Summary)
	return nil
}
