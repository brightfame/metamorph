package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
)

// OpenRepo returns a git.Repository object for the given repoPath.
func OpenRepo(repoPath string, tempDir string) (*git.Repository, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local repo: %s", err)
	}

	return repo, nil
}
