package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getkaze/helm/internal/config"
	"github.com/getkaze/helm/internal/display"
	"github.com/getkaze/helm/internal/session"
	"github.com/spf13/cobra"
)

var (
	initName  string
	initType  string
	initLang  string
	initForce bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Helm in the current directory",
	Long:  "Create the .helm/ directory structure, session.yaml, and helm.yaml.\nSupports interactive prompts and flag-based non-interactive mode.",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().StringVar(&initName, "name", "", "project name")
	initCmd.Flags().StringVar(&initType, "type", "", "project type (greenfield|brownfield)")
	initCmd.Flags().StringVar(&initLang, "lang", "", "language (en-US|pt-BR)")
	initCmd.Flags().BoolVar(&initForce, "force", false, "reinitialize without confirmation")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if .helm/ already exists
	if _, err := os.Stat(session.HelmDir); err == nil {
		if !initForce {
			if !isInteractive() {
				return fmt.Errorf("Helm already initialized. Use --force to reinitialize")
			}
			answer := display.Prompt("Helm already initialized. Reinitialize? (y/N)")
			if !strings.HasPrefix(strings.ToLower(answer), "y") {
				fmt.Fprintln(display.Out, "  Aborted.")
				return nil
			}
		}
	}

	interactive := isInteractive()

	// Resolve project name
	name := initName
	if name == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		defaultName := filepath.Base(cwd)
		if interactive {
			name = display.Prompt(fmt.Sprintf("Project name [%s]:", defaultName))
			if name == "" {
				name = defaultName
			}
		} else {
			name = defaultName
		}
	}

	// Resolve project type
	projectType := initType
	if projectType == "" {
		detected := detectProjectType()
		if interactive {
			answer := display.Prompt(fmt.Sprintf("Project type [%s] (greenfield|brownfield):", detected))
			if answer != "" {
				projectType = answer
			} else {
				projectType = detected
			}
		} else {
			projectType = detected
		}
	}
	if projectType != "greenfield" && projectType != "brownfield" {
		return fmt.Errorf("invalid project type %q: must be greenfield or brownfield", projectType)
	}

	// Resolve language
	lang := initLang
	if lang == "" {
		detected := detectLanguage()
		if interactive {
			answer := display.Prompt(fmt.Sprintf("Language [%s] (en-US|pt-BR):", detected))
			if answer != "" {
				lang = answer
			} else {
				lang = detected
			}
		} else {
			lang = detected
		}
	}
	if lang != "en-US" && lang != "pt-BR" {
		return fmt.Errorf("invalid language %q: must be en-US or pt-BR", lang)
	}

	// Create directories
	dirs := []string{
		session.HelmDir,
		session.HandoffsDir,
		session.ArtifactsDir,
		session.CheckpointsDir,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Extract embedded framework files (agents, rules, schemas, CLAUDE.md, .claude/)
	if err := extractEmbedded(); err != nil {
		return fmt.Errorf("failed to extract framework files: %w", err)
	}

	// Create session
	firstAgent := "scout"
	if projectType == "brownfield" {
		firstAgent = "survey"
	}

	s := &session.Session{
		Project: session.Project{
			Name:  name,
			Type:  projectType,
			State: "discover",
		},
		ExecutionProfile: "guided",
		CurrentAgent:     firstAgent,
		Language:         lang,
		Workflow:         "standard",
	}

	if err := session.Save(session.SessionFile, s); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Create config
	cfg := &config.Config{
		Version:     Version,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
		ProjectType: projectType,
		Language:    lang,
	}

	if err := config.Save(session.ConfigFile, cfg); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	// Check .gitignore
	checkGitignore()

	// Display summary
	display.InitSummary(name, projectType, lang)

	return nil
}

func detectProjectType() string {
	sourceExts := []string{".go", ".ts", ".py", ".rs", ".java", ".rb", ".js"}
	entries, err := os.ReadDir(".")
	if err != nil {
		return "greenfield"
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		for _, se := range sourceExts {
			if ext == se {
				return "brownfield"
			}
		}
	}
	return "greenfield"
}

func detectLanguage() string {
	for _, env := range []string{"LANG", "LC_ALL", "LANGUAGE"} {
		val := os.Getenv(env)
		if strings.Contains(strings.ToLower(val), "pt") {
			return "pt-BR"
		}
	}
	return "en-US"
}

func checkGitignore() {
	data, err := os.ReadFile(".gitignore")
	if err != nil {
		return
	}
	if !strings.Contains(string(data), ".helm/") && !strings.Contains(string(data), ".helm") {
		display.Warning(".helm/ is not in .gitignore. Runtime state may be committed.")
	}
}

func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func extractEmbedded() error {
	return fs.WalkDir(Embedded, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}

		if d.IsDir() {
			return os.MkdirAll(path, 0o755)
		}

		// Skip if file already exists (don't overwrite user changes)
		if _, err := os.Stat(path); err == nil {
			return nil
		}

		data, err := Embedded.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		return os.WriteFile(path, data, 0o644)
	})
}

