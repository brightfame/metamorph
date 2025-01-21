package git

import (
	"testing"
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
			},
			expectedError: false,
		},
		{
			name: "git basic clone",
			input: CloneOptions{
				URL:    "git@github.com:brightfame/metamorph-automated-tests.git",
				Branch: "",
			},
			expectedError: false,
		},
		{
			name: "https clone with branch",
			input: CloneOptions{
				URL:    "https://github.com/brightfame/metamorph-automated-tests",
				Branch: "branch-test",
			},
			expectedError: false,
		},
		{
			name: "git clone with branch",
			input: CloneOptions{
				URL:    "git@github.com:brightfame/metamorph-automated-tests.git",
				Branch: "branch-test",
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
