package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/getkaze/helm/internal/display"
	"github.com/spf13/cobra"
)

var Version = "dev"

var noColor bool

var rootCmd = &cobra.Command{
	Use:   "helm",
	Short: "AI agent orchestration for backend development",
	Long:  "Helm manages the lifecycle of AI agent pipeline sessions.\nUse it to initialize projects, monitor progress, checkpoint state, and resume work.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor || os.Getenv("NO_COLOR") != "" {
			color.NoColor = true
		}
	},
	SilenceUsage:         true,
	SilenceErrors:        true,
	CompletionOptions:    cobra.CompletionOptions{DisableDefaultCmd: true},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		display.Error(err.Error())
		fmt.Fprintln(os.Stderr, "Run 'helm --help' for usage.")
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
}

func errNoSession(err error) error {
	if os.IsNotExist(err) {
		return fmt.Errorf("No Helm session found. Run 'helm init' first.")
	}
	return err
}
