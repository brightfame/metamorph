package fileutil

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsSubFolder(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		parent   string
		child    string
		expected bool
	}{
		{"Parent and parent", "/parent", "/parent", true},
		{"Parent and child", "/parent", "/parent/child", true},
		{"Parent and grandchild", "/parent", "/parent/child/grandchild", true},
		{"Parent and sibling", "/parent", "/sibling", false},
		{"Child and parent", "/parent/child", "/parent", false},
		{"Child and sibling", "/parent/child", "/parent/childsibling", false},
	}

	fixtureRoot := "../test/fixtures/util"

	for _, testCase := range testCases {
		// The following is necessary to make sure testCase's values don't
		// get updated due to concurrency within the scope of t.Run(..) below
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			parent := filepath.Join(fixtureRoot, testCase.parent)
			child := filepath.Join(fixtureRoot, testCase.child)
			actual, err := IsSubFolder(parent, child)
			require.NoError(t, err)
			require.Equal(t, testCase.expected, actual)
		})
	}
}
