package shell

import (
	"bufio"
	"io"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Output contains the output after runnig a command.
type Output struct {
	stdout *outputStream
	stderr *outputStream
	// merged contains stdout  and stderr merged into one stream.
	merged *merged
}

func newOutput() *Output {
	m := new(merged)
	return &Output{
		merged: m,
		stdout: &outputStream{
			merged: m,
		},
		stderr: &outputStream{
			merged: m,
		},
	}
}

func (o *Output) Stdout() string {
	if o == nil {
		return ""
	}

	return o.stdout.String()
}

func (o *Output) Stderr() string {
	if o == nil {
		return ""
	}

	return o.stderr.String()
}

func (o *Output) Combined() string {
	if o == nil {
		return ""
	}

	return o.merged.String()
}

type outputStream struct {
	Lines []string
	*merged
}

func (st *outputStream) WriteString(s string) (n int, err error) {
	st.Lines = append(st.Lines, string(s))
	return st.merged.WriteString(s)
}

func (st *outputStream) String() string {
	if st == nil {
		return ""
	}

	return strings.Join(st.Lines, "")
}

type merged struct {
	// ensure that there are no parallel writes
	sync.Mutex
	Lines []string
}

func (m *merged) String() string {
	if m == nil {
		return ""
	}

	return strings.Join(m.Lines, "")
}

func (m *merged) WriteString(s string) (n int, err error) {
	m.Lock()
	defer m.Unlock()

	m.Lines = append(m.Lines, string(s))

	return len(s), nil
}

// This function captures stdout and stderr into the given variables while still printing it to the stdout and stderr
// of this Go program.
func readStdoutAndStderr(logger *zap.SugaredLogger, streamOutput bool, stdout, stderr io.ReadCloser) (*Output, error) {
	out := newOutput()
	stdoutReader := bufio.NewReader(stdout)
	stderrReader := bufio.NewReader(stderr)

	wg := &sync.WaitGroup{}

	wg.Add(2)
	var stdoutErr, stderrErr error
	go func() {
		defer wg.Done()
		stdoutErr = readData(logger, streamOutput, stdoutReader, out.stdout)
	}()
	go func() {
		defer wg.Done()
		stderrErr = readData(logger, streamOutput, stderrReader, out.stderr)
	}()
	wg.Wait()

	if stdoutErr != nil {
		return out, stdoutErr
	}
	if stderrErr != nil {
		return out, stderrErr
	}

	return out, nil
}

func readData(logger *zap.SugaredLogger, streamOutput bool, reader *bufio.Reader, writer io.StringWriter) error {
	var line string
	var readErr error
	for {
		line, readErr = reader.ReadString('\n')

		// only return early if the line does not have
		// any contents. We could have a line that does
		// not not have a newline before io.EOF, we still
		// need to add it to the output.
		if len(line) == 0 && readErr == io.EOF {
			break
		}

		if streamOutput {
			logger.Logln(zapcore.InfoLevel, line)
		}
		if _, err := writer.WriteString(line); err != nil {
			return err
		}

		if readErr != nil {
			break
		}
	}
	if readErr != io.EOF {
		return readErr
	}
	return nil
}
