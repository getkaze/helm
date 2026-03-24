package cmd

import (
	"fmt"

	"github.com/getkaze/helm/internal/display"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Helm version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(display.Out, "helm %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
