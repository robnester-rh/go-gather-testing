// Copyright The Enterprise Contract Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	gitMetadata "github.com/enterprise-contract/go-gather/metadata/git"
)

type MockSSHAuthenticator struct {
	mock.Mock
}

func (m *MockSSHAuthenticator) NewSSHAgentAuth(user string) (transport.AuthMethod, error) {
	args := m.Called(user)
	return nil, args.Error(1)
}

type MockMetadata struct {
	mock.Mock
}

func (m *MockMetadata) ForEach(fn func(*object.Commit) error) error {
	args := m.Called()
	return args.Error(0)
}

type MockCloner struct {
	mock.Mock
}

func (m *MockCloner) PlainClone(destination string, isBare bool, opts *git.CloneOptions) (*git.Repository, error) {
	args := m.Called(destination, isBare, opts)
	return args.Get(0).(*git.Repository), args.Error(1)
}

// MockRepository simulates a git repository for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CommitObjects() (object.CommitIter, error) {
	args := m.Called()
	return args.Get(0).(object.CommitIter), args.Error(1)
}

func TestGetGitCloneOptions_https_transport(t *testing.T) {
	srcURL := "git::https://github.com/example/repo.git"
	expectedCloneOpts := &git.CloneOptions{
		URL:   "https://github.com/example/repo.git",
		Depth: 1,
	}

	cloneOpts, err := getCloneOptions(srcURL, &RealSSHAuthenticator{})
	assert.NoError(t, err)
	assert.Equal(t, expectedCloneOpts, cloneOpts)
}

func TestGetGitCloneOptions_ssh_transport(t *testing.T) {
	srcURL := "git@github.com:example/repo.git"
	sshAuth, _ := ssh.NewSSHAgentAuth("git")
	expectedCloneOpts := &git.CloneOptions{
		URL:   "ssh://git@github.com/example/repo.git",
		Depth: 1,
		Auth:  sshAuth,
	}
	cloneOpts, err := getCloneOptions(srcURL, &RealSSHAuthenticator{})
	assert.NoError(t, err)
	assert.Equal(t, expectedCloneOpts.URL, cloneOpts.URL)
	assert.Equal(t, reflect.TypeOf(expectedCloneOpts.Auth), reflect.TypeOf(cloneOpts.Auth))
}

func TestGetGitCloneOptions_SSHAuthError(t *testing.T) {
	mockAuth := new(MockSSHAuthenticator)
	mockAuth.On("NewSSHAgentAuth", "git").Return(nil, fmt.Errorf("ssh auth error"))

	opts, err := getCloneOptions("ssh://example.com/repo.git", mockAuth)

	assert.Nil(t, opts)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "failed to create SSH auth method: ssh auth error")
	mockAuth.AssertExpectations(t)
}

// TestGatherSuccess tests the successful gathering of a git repository
func TestGatherSuccess(t *testing.T) {
	// Create a temporary directory for the repository
	dir, err := os.MkdirTemp("", "repo")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create a mock repository
	mockRepo := new(MockRepository)
	mockRepo.On("CommitObjects").Return(&MockMetadata{}, nil)

	// Create a mock cloner
	mockCloner := new(MockCloner)
	mockCloner.On("PlainClone", dir, false, &git.CloneOptions{}).Return(mockRepo, nil)

	// Create a mock authenticator
	mockAuth := new(MockSSHAuthenticator)
	mockAuth.On("NewSSHAgentAuth", "git").Return(nil, nil)

	// Create a gatherer with the mocks
	gatherer := &GitGatherer{}

	// Call the method under test
	ctx := context.Background()
	metadata, err := gatherer.Gather(ctx, "git::git@github.com:git-fixtures/basic.git", dir)

	// Assert that the metadata was returned
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
}

// TestGatherError_ProcessURL tests the error handling of the processURL function
func TestGatherError_ProcessURL(t *testing.T) {
	// Create a temporary directory for the repository
	dir, err := os.MkdirTemp("", "repo")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create a mock repository
	mockRepo := new(MockRepository)
	mockRepo.On("CommitObjects").Return(&MockMetadata{}, nil)

	// Create a mock cloner
	mockCloner := new(MockCloner)
	mockCloner.On("PlainClone", dir, false, &git.CloneOptions{}).Return(mockRepo, nil)

	// Create a mock authenticator
	mockAuth := new(MockSSHAuthenticator)
	mockAuth.On("NewSSHAgentAuth", "git").Return(nil, nil)

	// Create a gatherer with the mocks
	gatherer := &GitGatherer{}

	// Call the method under test
	ctx := context.Background()
	_, err = gatherer.Gather(ctx, "basic.git", dir)

	assert.EqualError(t, err, "failed to process URL: failed to classify URI: got basic.git. HTTP(S) URIs require a scheme (http:// or https://)")

}

// TestGather_CloneOpts tests that the ref and depth are set correctly in the clone options

func TestCloneRepositoryPath(t *testing.T) {
	// Create a temporary directory for the destination
	destination, err := os.MkdirTemp("", "test-destination")
	if err != nil {
		t.Fatal(err)
	}
	// defer os.RemoveAll(destination)

	// Create a temporary directory for the source repository
	sourceRepo, err := os.MkdirTemp("", "test-source-repo")
	if err != nil {
		t.Fatal(err)
	}
	// defer os.RemoveAll(sourceRepo)

	// Create a temporary directory for the subdirectory
	subdir, err := os.MkdirTemp(sourceRepo, "test-subdir")
	if err != nil {
		t.Fatal(err)
	}
	// defer os.RemoveAll(subdir)

	// Create a test file in the subdirectory
	testFile := filepath.Join(subdir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize a new Git repository in the source repository
	r, err := git.PlainInit(sourceRepo, false)
	if err != nil {
		t.Fatal(err)
	}

	// Add the test file to the Git repository
	w, err := r.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	_, err = w.Add(".")
	if err != nil {
		t.Fatal(err)
	}

	// Commit the changes to the Git repository
	commit, err := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create a new Git clone options
	cloneOpts := &git.CloneOptions{
		URL: sourceRepo,
	}

	// Clone the repository path
	metadata, err := cloneRepositoryPath(context.Background(), filepath.Base(subdir), destination, cloneOpts)
	if err != nil {
		t.Fatal(err)
	}

	// Assert that the destination directory exists
	_, err = os.Stat(destination)
	if err != nil {
		t.Fatalf("destination directory does not exist: %v", err)
	}

	// Assert that the test file exists in the destination directory
	_, err = os.Stat(filepath.Join(destination, "test.txt"))
	if err != nil {
		t.Fatalf("test file does not exist in the destination directory: %v", err)
	}

	// Assert that the metadata contains the commit
	gitMetadata, ok := metadata.(*gitMetadata.GitMetadata)
	if !ok {
		t.Fatalf("unexpected metadata type: %T", metadata)
	}
	if len(gitMetadata.Commits) != 1 {
		t.Fatalf("unexpected number of commits in metadata: %d", len(gitMetadata.Commits))
	}
	if gitMetadata.Commits[0].Hash.String() != commit.String() {
		t.Fatalf("unexpected commit hash in metadata: %s", gitMetadata.Commits[0].Hash.String())
	}
}
