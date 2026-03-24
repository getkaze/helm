package session

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// File system paths
const (
	HelmDir        = ".helm"
	SessionFile    = ".helm/session.yaml"
	SessionBackup  = ".helm/session.yaml.backup"
	ConfigFile     = "helm.yaml"
	HandoffsDir    = ".helm/handoffs"
	ArtifactsDir   = ".helm/artifacts"
	CheckpointsDir = ".helm/checkpoints"
)

// Pipeline order
var PipelineGreenfield = []string{"scout", "research", "planning", "architect", "roadmap", "breakdown", "review", "build", "verify", "ship"}
var PipelineBrownfield = []string{"survey", "research", "planning", "architect", "roadmap", "breakdown", "review", "build", "verify", "ship"}

// Valid enum values
var (
	ValidProjectTypes      = []string{"greenfield", "brownfield"}
	ValidProjectStates     = []string{"discover", "plan", "build", "validate", "deploy", "completed"}
	ValidAgentNames        = []string{"scout", "survey", "research", "planning", "architect", "roadmap", "breakdown", "review", "build", "verify", "ship"}
	ValidExecutionProfiles = []string{"explore", "guided", "autonomous"}
	ValidLanguages         = []string{"en-US", "pt-BR"}
	ValidWorkflows         = []string{"standard", "quick"}
	ValidAgentStatuses     = []string{"pending", "in_progress", "completed", "skipped", "needs_revalidation"}
)

type Session struct {
	Project          Project          `yaml:"project" json:"project"`
	ExecutionProfile string           `yaml:"execution_profile,omitempty" json:"execution_profile,omitempty"`
	CurrentAgent     string           `yaml:"current_agent" json:"current_agent"`
	Language         string           `yaml:"language" json:"language"`
	Workflow         string           `yaml:"workflow,omitempty" json:"workflow,omitempty"`
	InitialContext   string           `yaml:"initial_context,omitempty" json:"initial_context,omitempty"`
	Agents           map[string]Agent `yaml:"agents,omitempty" json:"agents,omitempty"`
	Tradeoffs        []Tradeoff       `yaml:"tradeoffs,omitempty" json:"tradeoffs,omitempty"`
	Deviations       []Deviation      `yaml:"deviations,omitempty" json:"deviations,omitempty"`
	LastCheckpoint   string           `yaml:"last_checkpoint,omitempty" json:"last_checkpoint,omitempty"`
}

type Project struct {
	Name  string `yaml:"name" json:"name"`
	Type  string `yaml:"type" json:"type"`
	State string `yaml:"state" json:"state"`
}

type Agent struct {
	Status        string `yaml:"status" json:"status"`
	Score         int    `yaml:"score,omitempty" json:"score,omitempty"`
	CriteriaCount int    `yaml:"criteria_count,omitempty" json:"criteria_count,omitempty"`
	CompletedAt   string `yaml:"completed_at,omitempty" json:"completed_at,omitempty"`
}

type Tradeoff struct {
	Decision  string `yaml:"decision" json:"decision"`
	Chosen    string `yaml:"chosen" json:"chosen"`
	Agent     string `yaml:"agent" json:"agent"`
	Artifact  string `yaml:"artifact,omitempty" json:"artifact,omitempty"`
	Timestamp string `yaml:"timestamp" json:"timestamp"`
}

type Deviation struct {
	Timestamp string `yaml:"timestamp" json:"timestamp"`
	Type      string `yaml:"type" json:"type"`
	FromAgent string `yaml:"from_agent,omitempty" json:"from_agent,omitempty"`
	ToAgent   string `yaml:"to_agent,omitempty" json:"to_agent,omitempty"`
	Reason    string `yaml:"reason" json:"reason"`
	Resolved  bool   `yaml:"resolved" json:"resolved"`
}

func Load(path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var s Session
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("session file is corrupted: %w", err)
	}

	return &s, nil
}

func Save(path string, s *Session) error {
	if err := Backup(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to backup session: %w", err)
	}

	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

func Backup(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	backupPath := path + ".backup"
	return os.WriteFile(backupPath, data, 0o644)
}

func Validate(s *Session) error {
	if s.Project.Name == "" {
		return fmt.Errorf("session missing required field: project.name")
	}
	if s.Project.Type == "" {
		return fmt.Errorf("session missing required field: project.type")
	}
	if !Contains(ValidProjectTypes, s.Project.Type) {
		return fmt.Errorf("invalid project.type %q: must be one of %v", s.Project.Type, ValidProjectTypes)
	}
	if s.Project.State == "" {
		return fmt.Errorf("session missing required field: project.state")
	}
	if !Contains(ValidProjectStates, s.Project.State) {
		return fmt.Errorf("invalid project.state %q: must be one of %v", s.Project.State, ValidProjectStates)
	}
	if s.CurrentAgent == "" {
		return fmt.Errorf("session missing required field: current_agent")
	}
	if !Contains(ValidAgentNames, s.CurrentAgent) {
		return fmt.Errorf("invalid current_agent %q: must be one of %v", s.CurrentAgent, ValidAgentNames)
	}
	if s.Language == "" {
		return fmt.Errorf("session missing required field: language")
	}
	if !Contains(ValidLanguages, s.Language) {
		return fmt.Errorf("invalid language %q: must be one of %v", s.Language, ValidLanguages)
	}
	if s.ExecutionProfile != "" && !Contains(ValidExecutionProfiles, s.ExecutionProfile) {
		return fmt.Errorf("invalid execution_profile %q: must be one of %v", s.ExecutionProfile, ValidExecutionProfiles)
	}
	if s.Workflow != "" && !Contains(ValidWorkflows, s.Workflow) {
		return fmt.Errorf("invalid workflow %q: must be one of %v", s.Workflow, ValidWorkflows)
	}

	for name, agent := range s.Agents {
		if !Contains(ValidAgentNames, name) {
			return fmt.Errorf("invalid agent name %q in agents map", name)
		}
		if !Contains(ValidAgentStatuses, agent.Status) {
			return fmt.Errorf("invalid status %q for agent %q: must be one of %v", agent.Status, name, ValidAgentStatuses)
		}
	}

	return nil
}

// Pipeline returns the pipeline order for the given project type.
func Pipeline(projectType string) []string {
	if projectType == "greenfield" {
		return PipelineGreenfield
	}
	return PipelineBrownfield
}

// Contains checks if a string slice contains a given item.
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
