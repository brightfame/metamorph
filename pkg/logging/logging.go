package logging

import (
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GetLogger returns a logger with the given log level and optionally color output disabled.
func GetLogger(out io.Writer, logLevel string, noColor bool) (*zap.SugaredLogger, error) {
	// Parse the log level
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}

	// Create a new Zap configuration
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	if noColor {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	cfg.EncoderConfig.EncodeCaller = nil
	cfg.DisableStacktrace = true
	if os.Getenv("METAMORPH_DEBUG") != "" {
		cfg.DisableStacktrace = false
		cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}
	cfg.Level = zap.NewAtomicLevelAt(level)

	// Create the console encoder with optional colored output
	consoleEncoder := zapcore.NewConsoleEncoder(cfg.EncoderConfig)
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(out), zap.NewAtomicLevelAt(level))
	core := zapcore.NewTee(consoleCore)

	// Build a new logger using the config
	logger, err := cfg.Build(zap.WrapCore(func(zapcore.Core) zapcore.Core {
		return core
	}))
	if err != nil {
		return nil, err
	}
	defer logger.Sync() //nolint:errcheck

	sugar := logger.Sugar().Named("metamorph")

	return sugar, nil
}
