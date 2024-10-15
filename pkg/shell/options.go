package shell

import (
	"go.uber.org/zap"
)

type ShellOptions struct {
	NonInteractive bool
	Logger         *zap.SugaredLogger
	WorkingDir     string
	SensitiveArgs  bool              // If true, will not log the arguments to the command
	Env            map[string]string // Additional environment variables to set
}

func NewShellOptions() *ShellOptions {
	logger, _ := zap.NewProduction()
	defer logger.Sync() //nolint:errcheck

	return &ShellOptions{
		NonInteractive: false,
		Logger:         logger.Sugar(),
		WorkingDir:     ".",
		SensitiveArgs:  false,
		Env:            map[string]string{},
	}
}
