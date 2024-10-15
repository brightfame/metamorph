package pipeline

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/brightfame/metamorph/internal/config"
)

type Pipeline struct {
	Name      string   `yaml:"name,omitempty"`
	WorkDir   string   `yaml:"work_dir,omitempty"`
	Assignees []string `yaml:"assignees,omitempty"`
	Reviewers []string `yaml:"reviewers,omitempty"`
	GitLab    GitLab   `yaml:"gitlab,omitempty"`
	Repos     []Repo   `yaml:"repos"`
	Steps     []Step   `json:"steps"`
	cfg       *config.Config
}

type GitLab struct {
	Org                     string   `yaml:"org"`
	BranchName              string   `yaml:"branch_name"`
	MergeRequestTitle       string   `yaml:"merge_request_title"`
	MergeRequestDescription string   `yaml:"merge_request_description"`
	Labels                  []string `yaml:"labels"`
}

type Repo struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type Step struct {
	Name     string            `yaml:"name,omitempty"`
	Image    string            `yaml:"image,omitempty"`
	Command  string            `yaml:"command,omitempty"`
	Env      map[string]string `yaml:"environment,omitempty"`
	WorkDir  string            `yaml:"work_dir,omitempty"`
	Volumes  []string          `yaml:"volumes,omitempty"`
	Timeout  string            `yaml:"timeout,omitempty"`
	Retry    RetryPolicy       `yaml:"retry,omitempty"`
	commands []string
}

// RetryPolicy defines the retry behavior for a step
type RetryPolicy struct {
	MaxAttempts int    `yaml:"max_attempts" json:"max_attempts"`
	Interval    string `yaml:"interval" json:"interval"`
}

// Commands returns the commands for the step
func (s *Step) Commands() []string {
	return s.commands
}

func New(cfg *config.Config, name string) *Pipeline {
	return &Pipeline{
		Name:  name,
		Steps: make([]Step, 0),
		cfg:   cfg,
	}
}

// LoadManifestFile loads the manifest from the given path and returns a pipeline.
// This also calls Validate() on the pipeline.
func LoadManifestFile(cfg *config.Config, path string) (*Pipeline, error) {
	p := &Pipeline{}
	p.cfg = cfg
	err := parseFile(p, path)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func parseFile(p *Pipeline, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer f.Close() // nolint: errcheck
	if err := yaml.NewDecoder(f).Decode(p); err != nil {
		return fmt.Errorf("decode config: %w", err)
	}

	// we map the command in each step to a slice of strings
	// so that we can pass it to the container executor
	for i, step := range p.Steps {
		p.Steps[i].commands = strings.Fields(step.Command)
	}

	return p.Validate()
}

func (p *Pipeline) AddStep(step Step) {
	p.Steps = append(p.Steps, step)
}

func (p *Pipeline) Validate() error {
	if len(p.Steps) == 0 {
		return fmt.Errorf("pipeline must contain at least one step")
	}
	for i, step := range p.Steps {
		if step.Name == "" {
			return fmt.Errorf("step %d must have a name", i)
		}
		if step.Image == "" {
			return fmt.Errorf("step %s must specify a Docker image", step.Name)
		}
		if len(step.commands) == 0 {
			return fmt.Errorf("step %s must specify at least one command", step.Name)
		}
	}
	return nil
}
