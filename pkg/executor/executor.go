package executor

import (
	"context"
	"io"
)

// ExecutionConfig holds the common configuration for any executor
type ExecutionConfig struct {
	WorkDir     string
	Environment map[string]string
	Command     []string
	Timeout     int
}

// ExecutionResult represents the output of an execution
type ExecutionResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Error    error
}

// Executor defines the interface for different execution environments
type Executor interface {
	// Initialize sets up the execution environment
	Initialize(ctx context.Context) error

	// Execute runs the payload and returns the result
	Execute(ctx context.Context, config ExecutionConfig) (*ExecutionResult, error)

	// Cleanup performs necessary cleanup after execution
	Cleanup(ctx context.Context) error

	// GetLogs retrieves execution logs
	GetLogs(ctx context.Context) (io.Reader, error)
}
