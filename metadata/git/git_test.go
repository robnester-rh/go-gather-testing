package git

import (
	"os"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

func TestGitMetadata_GetHashes(t *testing.T) {
	commits := []object.Commit{
		{Hash: plumbing.ComputeHash(plumbing.AnyObject, []byte("hash1"))},
		{Hash: plumbing.ComputeHash(plumbing.AnyObject, []byte("hash2"))},
		{Hash: plumbing.ComputeHash(plumbing.AnyObject, []byte("hash3"))},
	}

	metadata := GitMetadata{Commits: commits}
	expectedHashes := []string{"fc771c3730239d59dd35e5e0e1b527a78201d5fb", "3f655e38a3497759eba6493c02e0b8dcc7224243", "22e7c809dcfbc050a4b281c58f3da2959a3eeca9"}

	hashes := metadata.GetHashes()

	assert.Equal(t, expectedHashes, hashes)
}

func TestGitMetadata_Get(t *testing.T) {
	metadata := GitMetadata{
		Size:      100,
		Path:      "/path/to/repo",
		Timestamp: time.Now(),
		Commits: []object.Commit{
			{Hash: plumbing.ComputeHash(plumbing.AnyObject, []byte("hash1"))},
			{Hash: plumbing.ComputeHash(plumbing.AnyObject, []byte("hash2"))},
			{Hash: plumbing.ComputeHash(plumbing.AnyObject, []byte("hash3"))},
		},
	}

	expectedResult := map[string]any{
		"size":      int64(100),
		"path":      "/path/to/repo",
		"timestamp": metadata.Timestamp,
		"commits":   metadata.Commits,
	}

	defer os.RemoveAll(metadata.Path)
	result := metadata.Get()

	assert.Equal(t, expectedResult, result)
}
