package container

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/pkg/stdcopy"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	mmconfig "github.com/brightfame/metamorph/internal/config"
	"github.com/brightfame/metamorph/pkg/collections"
	"github.com/brightfame/metamorph/pkg/shell"
)

var (
	// ErrDockerRuntimeNotInstalled indicates the docker service is not installed or available in the path.
	ErrDockerRuntimeNotInstalled = errors.New("docker is not installed or available in the PATH. We recommend using Docker to isolate patch commands from your OS. Please install Docker or run MetaMorph using the --skip-container-runtime flag")
	// ErrDockerRuntimeNotStarted indicates the docker service hasn't been started.
	ErrDockerRuntimeNotStarted = errors.New("docker is not running. We recommend using Docker to isolate patch commands from your OS. Please start the docker service or run MetaMorph using the --skip-container-runtime flag")
)

// DockerRuntime represents the Docker container runtime.
type DockerRuntime struct {
	client *client.Client
	cfg    *mmconfig.Config
}

// NewDockerRuntime creates a new instance of the Docker runtime.
func NewDockerRuntime(cfg *mmconfig.Config) (*DockerRuntime, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerRuntime{
		client: cli,
		cfg:    cfg,
	}, nil
}

// Type returns the Docker runtime type.
func (d *DockerRuntime) Type() RuntimeType {
	return DockerRuntimeType
}

// IsAvailable returns an error when Docker isn't installed or available.
func (d *DockerRuntime) IsAvailable() error {
	if !shell.CommandInstalled("docker") {
		return ErrDockerRuntimeNotInstalled
	}

	cmd := exec.Command("docker", "info")
	if err := cmd.Start(); err != nil {
		return ErrDockerRuntimeNotStarted
	}
	if err := cmd.Wait(); err != nil {
		return ErrDockerRuntimeNotStarted
	}

	return nil
}

// BuildImage builds the desired image using contextDir as the source.
func (d *DockerRuntime) BuildImage(ctx context.Context, image DockerImage, contextDir string) (string, error) {
	// Set the build options
	buildOptions := types.ImageBuildOptions{
		Dockerfile: defaultDockerfile,
		Tags:       []string{image.String()},
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}

	excludes, err := build.ReadDockerignore(contextDir)
	if err != nil {
		return "", fmt.Errorf("unable to read .dockerignore: '%s'", err.Error())
	}

	if err := build.ValidateContextDirectory(contextDir, excludes); err != nil {
		return "", fmt.Errorf("error checking context: '%s'", err.Error())
	}

	excludes = build.TrimBuildFilesFromExcludes(excludes, defaultDockerfile, false)

	buildContext, err := archive.TarWithOptions(contextDir, &archive.TarOptions{
		ExcludePatterns: excludes,
		ChownOpts:       &idtools.Identity{UID: 0, GID: 0},
	})
	if err != nil {
		return "", fmt.Errorf("unable to compress context: '%s'", err.Error())
	}

	resp, err := cli.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return "", fmt.Errorf("could not build image, got error '%s'", err.Error())
	}

	defer resp.Body.Close()
	respBuffer, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read build image response, got error '%s'", err.Error())
	}

	d.cfg.Logger.Debug(string(respBuffer))
	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not copy the response body to stdout, got error '%s'", err.Error())
	}
	return string(respBuffer), nil
}

// PullImage ensures the required Docker image is available.
func (d *DockerRuntime) PullImage(ctx context.Context, img DockerImage) error {
	imageFound, err := d.findLocalImage(ctx, img)
	if err != nil {
		return err
	}

	if imageFound {
		return ErrImageExists
	}

	// Configure Docker registry auth if required
	pullOptions := image.PullOptions{}
	if d.cfg.PlatformAuthConfig.Username != "" && d.cfg.PlatformAuthConfig.Password != "" {
		authConfig := registry.AuthConfig{
			Username: d.cfg.PlatformAuthConfig.Username,
			Password: d.cfg.PlatformAuthConfig.Password,
		}
		authStr, err := registry.EncodeAuthConfig(authConfig)
		if err != nil {
			d.cfg.Logger.Error("unable to encode auth config", "error", err)
			return err
		}
		pullOptions.RegistryAuth = authStr
	}

	// Attempt to pull the image
	d.cfg.Logger.Infof("Pulling image %s", img.String())
	out, err := d.client.ImagePull(ctx, img.String(), pullOptions)

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		d.cfg.Logger.Error("timeout pulling container", "image_ref", img.String())
		return err
	}

	if err != nil {
		d.cfg.Logger.Error("failed pulling container", "image_ref", img.String(), "error", err)
		return err
	}

	defer out.Close()

	var dst bytes.Buffer
	_, err = io.Copy(&dst, out)
	if err != nil {
		return err
	}

	return nil
}

// Run creates and starts a Docker container with the specified configuration.
func (d *DockerRuntime) Run(ctx context.Context, containerID string, config *Config, hostConfig *HostConfig) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		d.cfg.Logger.Error(err)
		return err
	}

	// prepare commands
	cmds := config.Cmd
	if len(config.Cmd) > 0 {
		cmds = append([]string{"/bin/bash", "-c"}, config.Cmd...)
	}

	// prepare the Docker configuration
	dockerContainerConfig := container.Config{
		Image:        config.Image.String(),
		Cmd:          cmds,
		Entrypoint:   config.Entrypoint,
		Tty:          config.Tty,
		WorkingDir:   config.WorkingDir,
		AttachStderr: config.AttachStderr,
		AttachStdout: config.AttachStdout,
		Env:          collections.KeyValueStringSlice(config.Env),
	}

	mounts := make([]mount.Mount, 0)
	if hostConfig != nil {
		for _, m := range hostConfig.Mounts {
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: m.Source,
				Target: m.Target,
			})
		}
	}

	dockerHostConfig := container.HostConfig{
		Mounts: mounts,
	}

	// create the container
	resp, err := cli.ContainerCreate(ctx, &dockerContainerConfig, &dockerHostConfig, nil, nil, containerID)
	if err != nil {
		d.cfg.Logger.Error(err)
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		d.cfg.Logger.Error(err)
		return err
	}

	exitCode := -1
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			d.cfg.Logger.Error(err)
			return fmt.Errorf("docker runtime error while waiting for container '%s' to exit: %w", resp.ID, err)
		}
	case status := <-statusCh:
		exitCode = int(status.StatusCode)
	}

	// copy any logs from the container to the sugared logger instance
	out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	})
	if err != nil {
		return err
	}

	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(io.Writer(&stdout), io.Writer(&stderr), out)
	if err != nil {
		return err
	}

	// write the contents of each buffer to the logger. Note: we deliberately write to InfoLevel as the container
	// commands are not errors even though they may be written to stderr.
	copyBufToLogger(stdout, d.cfg.Logger, zap.InfoLevel)
	copyBufToLogger(stderr, d.cfg.Logger, zap.InfoLevel)

	if exitCode != 0 {
		return fmt.Errorf("script exited with status code %d", exitCode)
	}

	return nil
}

func copyBufToLogger(buf bytes.Buffer, logger *zap.SugaredLogger, level zapcore.Level) {
	// Loop over stdout and log each line
	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		line := scanner.Text()
		logger.Log(level, line)
	}
}

func (dr *DockerRuntime) findLocalImage(ctx context.Context, img DockerImage) (bool, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, err
	}

	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return false, err
	}

	for _, image := range images {
		if len(image.RepoTags) == 0 {
			continue
		}

		if img.String() == image.RepoTags[0] {
			return true, nil
		}
	}
	return false, nil
}
