package container

import (
	"context"
	"errors"
	"fmt"

	"github.com/brightfame/metamorph/internal/config"
)

// RuntimeType is the type of container runtime.
type RuntimeType string

const (
	// DockerRuntimeType is the Docker runtime type.
	DockerRuntimeType RuntimeType = "docker"
)

// String returns the runtime type string.
func (rt RuntimeType) String() string {
	return string(rt)
}

// ParseRuntimeType parses the given string into a RuntimeType.
func ParseRuntimeType(rt string) (RuntimeType, error) {
	switch rt {
	case DockerRuntimeType.String():
		return DockerRuntimeType, nil
	default:
		return "", fmt.Errorf("unknown runtime type: %s", rt)
	}
}

var (
	// ErrImageExists indicates that an image already exists.
	ErrImageExists = errors.New("container image: already exists")
	// ErrInvalidName indicates that a image name format is invalid.
	ErrInvalidName = errors.New("container image: invalid format. Should be <IMAGE_NAME>:<IMAGE_TAG>")
	// defaultDockerfile is the name of the default Dockerfile to look for
	defaultDockerfile = "Dockerfile"
)

// Runtime interface defines the interfaces that should be implemented by a container runtime.
type Runtime interface {
	// Type returns the type of the container runtime.
	Type() RuntimeType

	// IsAvailable returns an error if the runtime isn't installed or available.
	IsAvailable() error

	// BuildImage builds the desired image using contextDir as the source
	BuildImage(ctx context.Context, img DockerImage, contextDir string) (string, error)

	// PullImage pulls an image from the network to local storage. It returns any errors that occur.
	PullImage(ctx context.Context, img DockerImage) error

	// Run synchronously executes the command using the runtime, and returns any errors that occur.
	// If the command completes with a non-0 exit code, a ExitError will be returned.
	Run(ctx context.Context, containerID string, config *Config, hostConfig *HostConfig) error
}

// NewRuntime creates a new runtime instance using the specified type.
func NewRuntime(rt RuntimeType, cfg *config.Config) (Runtime, error) {
	if rt == DockerRuntimeType {
		return NewDockerRuntime(cfg)
	}

	return nil, fmt.Errorf("unknown container runtime: %s", rt)
}

// Config contains the configuration data about a container.
type Config struct {
	Image        DockerImage       // Name of the image
	Entrypoint   []string          // Entrypoint to run when starting the container
	Cmd          []string          // Command(s) to run when starting the container
	Tty          bool              // Attach standard streams to a tty, including stdin if it is not closed.
	WorkingDir   string            // Current directory (PWD) in the command will be launched
	AttachStdout bool              // Attach the standard output
	AttachStderr bool              // Attach the standard error
	Env          map[string]string // List of environment variables to set in the container
}

// HostConfig the non-portable Config structure of a container that is dependent of the host we are running on.
type HostConfig struct {
	// Mounts used by the container
	Mounts []Mount
}

// Mount represents a mount (volume).
type Mount struct {
	Source string // Source specifies the name of the mount.
	Target string // Target is the path within the container.
}
