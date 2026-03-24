package cmd

import (
	"github.com/getkaze/helm/internal/config"
	"github.com/getkaze/helm/internal/display"
	"github.com/getkaze/helm/internal/session"
	"github.com/spf13/cobra"
)

var (
	statusJSON  bool
	statusShort bool
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show pipeline status dashboard",
	Long:  "Display the current state of the project pipeline — all agents with status, score, and timestamps.",
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "output as JSON")
	statusCmd.Flags().BoolVar(&statusShort, "short", false, "one-line summary")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	s, err := session.Load(session.SessionFile)
	if err != nil {
		return errNoSession(err)
	}

	if statusJSON {
		return display.JSON(s)
	}

	if statusShort {
		display.Short(s)
		return nil
	}

	cfg, err := config.Load(session.ConfigFile)
	if err != nil {
		display.Warning("Could not load helm.yaml. Version info may be missing.")
	}
	display.Dashboard(s, cfg)
	return nil
}
