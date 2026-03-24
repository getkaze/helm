package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getkaze/helm/internal/checkpoint"
	"github.com/getkaze/helm/internal/display"
	"github.com/getkaze/helm/internal/session"
	"github.com/spf13/cobra"
)

var (
	saveForce   bool
	saveMessage string
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Checkpoint session state",
	Long:  "Validate session integrity and create a checkpoint for safe session handoff.",
	RunE:  runSave,
}

func init() {
	saveCmd.Flags().BoolVar(&saveForce, "force", false, "skip recent checkpoint confirmation")
	saveCmd.Flags().StringVar(&saveMessage, "message", "", "checkpoint message")
	rootCmd.AddCommand(saveCmd)
}

func runSave(cmd *cobra.Command, args []string) error {
	s, err := session.Load(session.SessionFile)
	if err != nil {
		return errNoSession(err)
	}

	// Validate session
	if err := session.Validate(s); err != nil {
		return fmt.Errorf("session validation failed: %w", err)
	}

	// Validate handoffs and artifacts for completed agents
	var warnings []string
	handoffCount := 0
	artifactCount := 0

	for name, agent := range s.Agents {
		if agent.Status != "completed" {
			continue
		}

		// Check handoff
		handoffPath := filepath.Join(session.HandoffsDir, name+".md")
		if _, err := os.Stat(handoffPath); err == nil {
			handoffCount++
		} else {
			warnings = append(warnings, fmt.Sprintf("Handoff for %s not found.", name))
		}

		// Check artifact directory
		artifactPath := filepath.Join(session.ArtifactsDir, name)
		if info, err := os.Stat(artifactPath); err == nil && info.IsDir() {
			artifactCount++
		} else {
			warnings = append(warnings, fmt.Sprintf("Artifact directory for %s not found.", name))
		}
	}

	// Check for recent checkpoint
	if !saveForce {
		_, latestTime, _ := checkpoint.LatestTimestamp(session.CheckpointsDir)
		if !latestTime.IsZero() && time.Since(latestTime) < 5*time.Minute {
			ts := latestTime.Format("15:04:05")
			if !isInteractive() {
				return fmt.Errorf("recent checkpoint exists (%s). Use --force to overwrite", ts)
			}
			answer := display.Prompt(fmt.Sprintf("Recent checkpoint exists (%s). Overwrite? (y/N)", ts))
			if !strings.HasPrefix(strings.ToLower(answer), "y") {
				fmt.Fprintln(display.Out, "  Aborted.")
				return nil
			}
		}
	}

	// Backup session
	if err := session.Backup(session.SessionFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to backup session: %w", err)
	}

	// Rotate before creating new checkpoint
	if err := checkpoint.Rotate(session.CheckpointsDir, 5); err != nil {
		return fmt.Errorf("failed to rotate checkpoints: %w", err)
	}

	// Create checkpoint
	ts, err := checkpoint.Create(session.CheckpointsDir, s, saveMessage)
	if err != nil {
		return fmt.Errorf("failed to create checkpoint: %w", err)
	}

	// Update last_checkpoint in session
	s.LastCheckpoint = ts
	if err := session.Save(session.SessionFile, s); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	display.SaveConfirmation(ts, s, handoffCount, artifactCount, warnings)
	return nil
}
