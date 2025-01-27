package git

import (
	"os"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func TestGitOperations(t *testing.T) {
	tests := []struct {
		name          string
		input         CloneOptions
		expectedError bool
	}{
		{
			name: "https basic clone",
			input: CloneOptions{
				URL:    "https://github.com/brightfame/metamorph-automated-tests",
				Branch: "",
				Auth: &http.BasicAuth{
					Username: "metamorph",
					Password: os.Getenv("GITHUB_OAUTH_TOKEN"),
				},
			},
			expectedError: false,
		},
		{
			name: "git basic clone",
			input: CloneOptions{
				URL:    "git@github.com:brightfame/metamorph-automated-tests.git",
				Branch: "",
				Auth:   sshAgentAuth(),
			},
			expectedError: false,
		},
		{
			name: "https clone with branch",
			input: CloneOptions{
				URL:    "https://github.com/brightfame/metamorph-automated-tests",
				Branch: "branch-test",
				Auth: &http.BasicAuth{
					Username: "metamorph",
					Password: os.Getenv("GITHUB_OAUTH_TOKEN"),
				},
			},
			expectedError: false,
		},
		{
			name: "git clone with branch",
			input: CloneOptions{
				URL:    "git@github.com:brightfame/metamorph-automated-tests.git",
				Branch: "branch-test",
				Auth:   sshAgentAuth(),
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a temp directory for the clone
			tt.input.Destination = t.TempDir()

			err := Clone(tt.input)
			if tt.expectedError && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func sshAgentAuth() *ssh.PublicKeysCallback {
	auth, err := ssh.NewSSHAgentAuth("git")
	if err != nil {
		panic(err)
	}

	return auth
}
