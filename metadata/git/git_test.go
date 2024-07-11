package git

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
)

func TestGitMetadata_Get(t *testing.T) {
	metadata := GitMetadata{
		LatestCommit: plumbing.ComputeHash(plumbing.AnyObject, []byte("hash1")).String(),
	}

	expectedResult := map[string]any{
		"latest_commit":   metadata.LatestCommit,
	}
	result := metadata.Get()

	assert.Equal(t, expectedResult, result)
}


// TestGetLatestCommit tests the GetLatestCommit method.
func TestGitMetadata_GetLatestCommit(t *testing.T) {
	metadata := GitMetadata{
		LatestCommit: plumbing.ComputeHash(plumbing.AnyObject, []byte("hash1")).String(),
	}

	expectedResult := metadata.LatestCommit
	result := metadata.GetLatestCommit()

	assert.Equal(t, expectedResult, result)
}