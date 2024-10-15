package runner

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/brightfame/metamorph/internal/config"
	"github.com/brightfame/metamorph/internal/fileutil"
	"github.com/brightfame/metamorph/pkg/container"
	"github.com/brightfame/metamorph/pkg/pipeline"
)

// Runner executes a pipeline of tasks
type Runner struct {
	p        *pipeline.Pipeline
	errChan  chan error
	doneChan chan bool
	mutex    sync.Mutex
	cr       container.Runtime
	cfg      *config.Config
}

type Result struct {
	StepName string
	ExitCode int
	Output   []byte
	Error    error
	Duration time.Duration
}

// New creates a new Runner instance
func New(cfg *config.Config, p *pipeline.Pipeline) *Runner {
	return &Runner{
		p:        p,
		errChan:  make(chan error, 1),
		doneChan: make(chan bool, 1),
		cfg:      cfg,
	}
}

// Run executes all steps in the pipeline
func (r *Runner) Run(ctx context.Context) ([]Result, error) {
	// create the container runtime instance
	rt, err := container.ParseRuntimeType(r.cfg.ContainerRuntime)
	if err != nil {
		return nil, err
	}

	runtime, err := container.NewRuntime(rt, r.cfg)
	if err != nil {
		return nil, err
	}
	r.cr = runtime

	r.cfg.Logger.Infof("Starting pipeline execution", "steps", len(r.p.Steps))

	results := make([]Result, 0, len(r.p.Steps))

	// execute the pipeline for each repo
	for _, repo := range r.p.Repos {
		repoLogger := r.cfg.Logger.With("repo", repo.Name)
		repoLogger.Info("Starting pipeline execution for %s", repo.Name)

		for i, step := range r.p.Steps {
			stepLogger := repoLogger.With("step", step.Name, "step_number", i+1)
			stepLogger.Infof("Executing", "commands", step.Commands())

			// set defaults
			if step.WorkDir == "" {
				step.WorkDir = r.p.WorkDir
			}

			select {
			case <-ctx.Done():
				return results, ctx.Err()
			default:
				// execute the step
				start := time.Now()
				result, err := r.executeStepImpl(ctx, repo, step, stepLogger)
				duration := time.Since(start)

				if err != nil {
					return results, fmt.Errorf("step execution failed: %w", err)
				}

				stepLogger.Infof("Step completed successfully", "duration", duration, "exit_code", result.ExitCode)

				results = append(results, result)
			}
		}
	}

	return results, nil
}

func (r *Runner) executeStepImpl(ctx context.Context, repo pipeline.Repo, step pipeline.Step, logger *zap.SugaredLogger) (Result, error) {
	image := container.ParseDockerImage(step.Image)

	// ensure the container image exists and pull it if necessary
	err := r.cr.PullImage(ctx, image)
	if errors.Is(err, container.ErrImageExists) {
		logger.Debugf("Image %s already exists.", image)
	} else if err != nil {
		return Result{}, err
	}

	cConfig := &container.Config{
		Image:        image,
		Cmd:          step.Commands(),
		Tty:          false,
		WorkingDir:   r.cfg.DefaultContainerRepoPath,
		AttachStdout: true,
		AttachStderr: true,
		Env:          step.Env,
	}

	// process any volume mounts
	mounts := make([]container.Mount, 0)
	for _, volume := range step.Volumes {
		// volumes are stored in the format source:target, so we need to split them
		mount := strings.Split(volume, ":")

		// ensure the source path is absolute
		sourceAbs, err := filepath.Abs(mount[0])
		if err != nil {
			return Result{}, err
		}

		mounts = append(mounts, container.Mount{
			Source: sourceAbs,
			Target: mount[1],
		})
	}

	// explicitly add a mount for the repo root
	repoRootPath := fileutil.RepoRootPath(repo.Path, logger)
	mounts = append(mounts, container.Mount{
		Source: repoRootPath,
		Target: r.cfg.DefaultContainerRepoPath,
	})

	hostConfig := &container.HostConfig{
		Mounts: mounts,
	}

	if err := r.cr.Run(ctx, "", cConfig, hostConfig); err != nil {
		return Result{}, err
	}

	return Result{
		StepName: step.Name,
		ExitCode: 0,
		Output:   nil,
		Error:    nil,
		Duration: 0,
	}, nil
}

// RunAsync executes all steps in the pipeline asynchronously
// func (r *Runner) RunAsync(ctx context.Context) {
// 	go func() {
// 		if err := r.Run(ctx); err != nil {
// 			r.errChan <- err
// 			return
// 		}
// 		r.doneChan <- true
// 	}()
// }

// Wait waits for the pipeline to complete and returns any error
func (r *Runner) Wait() error {
	select {
	case err := <-r.errChan:
		return err
	case <-r.doneChan:
		return nil
	}
}
