package config

import (
	"os"

	"go.uber.org/zap"

	"github.com/brightfame/metamorph/pkg/logging"
)

// A Config implements persistent storage and modification of application configuration.
type Config struct {
	// Logger is the default Logger instance.
	Logger *zap.SugaredLogger
	// WorkingDir is the path to the working directory.
	WorkingDir string
	// DefaultContainerRepoPath is the path inside the container to mount the repository.
	DefaultContainerRepoPath string
	// Repos is a list of repositories to work with.
	Repos []string `yaml:"repos"`
	// Platform is the SCM platform you are working with.
	Platform           string             `yaml:"platform,omitempty"`
	PlatformOrg        string             `yaml:"platform_org,omitempty"`
	PlatformAuthConfig PlatformAuthConfig `yaml:"platform_auth_config,omitempty"`
	// ContainerRuntime is the container runtime to use.
	ContainerRuntime string `yaml:"container_runtime,omitempty"`
}

// PlatformAuthConfig is the authentication configuration for the SCM platform.
type PlatformAuthConfig struct {
	Host     string // e.g. "registry.gitlab.com"
	Username string
	Password string // or Token
}

// DefaultConfig returns the default Config. All the path values are relative
// to the data directory.
// Use Validate() to validate the config and ensure absolute paths.
func DefaultConfig() (*Config, error) {
	logger, err := logging.GetLogger(os.Stdout, "info", false)
	if err != nil {
		return nil, err
	}

	// try to compute the repo root path
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return &Config{
		Logger:                   logger,
		WorkingDir:               workingDir,
		DefaultContainerRepoPath: "/usr/src/repo",
		Platform:                 "gitlab",
		PlatformAuthConfig: PlatformAuthConfig{
			Host:     "registry.gitlab.com",
			Username: "",
			Password: "",
		},
		ContainerRuntime: "docker",
	}, nil
}

// Validate validates the configuration.
// It updates the configuration with absolute paths.
func (c *Config) Validate() error {
	return nil
}
