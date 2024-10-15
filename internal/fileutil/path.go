package fileutil

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// GetPathRelativeTo returns the relative path you would have to take to get from basePath to path. If either path
// or basePath are symbolic links, they will be evaluated before the relative path between them is calculated.
func GetPathRelativeTo(path string, basePath string) (string, error) {
	if path == "" {
		path = "."
	}
	if basePath == "" {
		basePath = "."
	}

	basePathEvaluated, err := filepath.EvalSymlinks(basePath)
	if err != nil {
		return "", err
	}

	inputFolderAbs, err := filepath.Abs(basePathEvaluated)
	if err != nil {
		return "", err
	}

	pathEvaluated, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}

	fileAbs, err := filepath.Abs(pathEvaluated)
	if err != nil {
		return "", err
	}

	relPath, err := filepath.Rel(inputFolderAbs, fileAbs)
	if err != nil {
		return "", err
	}

	return filepath.ToSlash(relPath), nil
}

// RepoRootPath returns the Git repo root in the given workingDir or any of its parent dirs, assuming workingDir is in
// a Git repository. If it's not, returns the given workingDir.
func RepoRootPath(workingDir string, logger *zap.SugaredLogger) string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = workingDir
	path, err := cmd.Output()
	if err != nil {
		// The git command exits with 128 code when the current folder is not a Git project
		// the function then returns the current dir
		logger.Debugf("git rev-parse returned an error: %v. Falling back to working dir: %s", err, workingDir)
		return workingDir
	}

	return strings.TrimSpace(string(path))
}

// EnsureDirectoriesEqual diffs two directories to ensure they have the exact same files, contents, etc and showing exactly what's different.
// Hidden files/directories are ignored.
// TODO: this function will break on Windows.
func EnsureDirectoriesEqual(folderWithExpectedContents string, folderWithActualContents string) error {
	cmd := exec.Command("diff", "-x", ".*", "-r", "-u", folderWithExpectedContents, folderWithActualContents)

	bytes, err := cmd.Output()
	output := string(bytes)

	if err != nil {
		return fmt.Errorf("diff command exited with an error. This likely means the contents of %s and %s are different. Here is the output of the diff command:\n%s", folderWithExpectedContents, folderWithActualContents, output)
	}

	return nil
}

func CompareDirectories(directory1 string, directory2 string) (string, error) {
	cmd := exec.Command("diff", "-r", directory1, directory2)

	bytes, err := cmd.Output()
	output := string(bytes)

	// The diff command exits 1 when there are differences found
	if cmd.ProcessState.ExitCode() > 1 {
		return "", err
	}

	return output, nil
}

// IsSubFolder returns true if child is a sub folder of parent, and false otherwise.
// Solution inspired by: https://stackoverflow.com/questions/28024731/check-if-given-path-is-a-subdirectory-of-another-in-golang
func IsSubFolder(parent, child string) (bool, error) {
	parentEvaluated, err := filepath.EvalSymlinks(parent)
	if err != nil {
		return false, err
	}

	parentAbs, err := filepath.Abs(parentEvaluated)
	if err != nil {
		return false, err
	}

	childEvaluated, err := filepath.EvalSymlinks(child)
	if err != nil {
		return false, err
	}

	childAbs, err := filepath.Abs(childEvaluated)
	if err != nil {
		return false, err
	}

	relPath, err := filepath.Rel(parentAbs, childAbs)
	if err != nil {
		return false, err
	}

	// If relative path between parent and child is the current directory, then parent and child are
	// the same folder, which is allowed
	if relPath == "." {
		return true, nil
	}

	// Relative path between parent and child must not contain any ".."
	if strings.Contains(relPath, "..") {
		return false, nil
	}

	// Relative path between parent and child must be completely included in the child
	if !strings.Contains(childAbs, relPath) {
		return false, nil
	}

	return true, nil
}
