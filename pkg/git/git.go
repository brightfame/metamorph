package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

type CloneOptions struct {
	URL         string
	Branch      string
	Destination string
	Auth        transport.AuthMethod
}

func Clone(opts CloneOptions) error {
	cloneOpts := &git.CloneOptions{
		URL: opts.URL,
	}

	if opts.Auth != nil {
		cloneOpts.Auth = opts.Auth
	}

	if opts.Branch != "" {
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(opts.Branch)
		cloneOpts.SingleBranch = true
	}

	_, err := git.PlainClone(opts.Destination, false, cloneOpts)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}
