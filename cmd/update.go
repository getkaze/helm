package cmd

import (
	"fmt"
	"os"

	"github.com/getkaze/helm/internal/display"
	"github.com/getkaze/helm/internal/updater"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Helm to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		runUpdate(Version)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(current string) {
	fmt.Fprintf(display.Out, "  Checking for updates...\n")

	result, err := updater.Check(current)
	if err != nil {
		display.Error(fmt.Sprintf("check failed: %v", err))
		os.Exit(1)
	}

	if !result.Available {
		fmt.Fprintf(display.Out, "  Already up to date (%s).\n", current)
		return
	}

	fmt.Fprintf(display.Out, "  New version available: %s (current: %s)\n", result.Latest, current)
	fmt.Fprintf(display.Out, "  Downloading helm %s...\n", result.Latest)

	tmpPath, err := updater.Download(result.Latest)
	if err != nil {
		display.Error(fmt.Sprintf("download failed: %v", err))
		os.Exit(1)
	}
	defer os.Remove(tmpPath)

	if err := updater.Replace(tmpPath); err != nil {
		display.Error(fmt.Sprintf("replace failed: %v\n\nhint: try running with sudo", err))
		os.Exit(1)
	}

	fmt.Fprintf(display.Out, "  Updated to %s.\n", result.Latest)
}
