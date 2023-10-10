package logging

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetContainerLogger(t *testing.T) {
	logMessage := "Testing logging for the container. This line should appear in a file."

	logger, logFilePath, err := GetLoggerWithNewFile(io.Discard, "info", true)
	require.NoError(t, err)

	logger.Infof(logMessage)

	// test if the log file exists
	assert.FileExists(t, logFilePath)
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Error("Failed to read the file data")
	}

	defer os.Remove(logFilePath)

	// check contents
	assert.Contains(t, string(content), logMessage)

	// cleanup
	require.NoError(t, err, "Failed to remove test log file. Please remove manually for now: %v", logFilePath)
}
