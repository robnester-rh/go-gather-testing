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
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (m *MockCloner) PlainCloneContext(ctx context.Context, destination string, isBare bool, opts *git.CloneOptions) (*git.Repository, error) {
	args := m.Called(ctx, destination, isBare, opts)
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

// TestGatherSuccess_withSubDir tests the successful gathering of a git repository with a subdirectory
func TestGatherSuccess_withSubDir(t *testing.T) {
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
	mockCloner.On("PlainCloneContext", mock.Anything, false, &git.CloneOptions{URL: "https://github.com/git-fixtures/basic.git"}).Return(mockRepo, nil)

	// Create a mock authenticator
	mockAuth := new(MockSSHAuthenticator)
	mockAuth.On("NewSSHAgentAuth", "git").Return(nil, nil)

	// Create a gatherer with the mocks
	gatherer := &GitGatherer{}

	// Call the method under test
	ctx := context.Background()
	metadata, err := gatherer.Gather(ctx, "https://github.com/git-fixtures/basic.git//go", dir)

	// Assert that the metadata was returned
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
}

// TestGatherSuccess_withRef tests the successful gathering of a git repository with a ref
func TestGatherSuccess_withRef(t *testing.T) {
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
	mockCloner.On("PlainClone", dir, false, &git.CloneOptions{URL: "https://github.com/git-fixtures/basic.git", Depth: 1, ReferenceName: "refs/heads/main"}).Return(mockRepo, nil)

	// Create a mock authenticator
	mockAuth := new(MockSSHAuthenticator)
	mockAuth.On("NewSSHAgentAuth", "git").Return(nil, nil)

	// Create a gatherer with the mocks
	gatherer := &GitGatherer{}

	// Call the method under test
	ctx := context.Background()
	metadata, err := gatherer.Gather(ctx, "https://github.com/git-fixtures/basic.git?ref=main", dir)

	// Assert that the metadata was returned
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
}

// TestGatherError_ParseDepth tests the error handling when depth is invalid
func TestGatherError_ParseDepth(t *testing.T) {
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
	mockCloner.On("PlainClone", dir, false, &git.CloneOptions{URL: "https://github.com/git-fixtures/basic.git?depth=squiggle"}).Return(mockRepo, nil)

	// Create a mock authenticator
	mockAuth := new(MockSSHAuthenticator)
	mockAuth.On("NewSSHAgentAuth", "git").Return(nil, nil)

	// Create a gatherer with the mocks
	gatherer := &GitGatherer{}

	// Call the method under test
	ctx := context.Background()
	_, err = gatherer.Gather(ctx, "https://github.com/git-fixtures/basic.git?depth=squiggle", dir)

	assert.EqualError(t, err, "failed to parse depth: strconv.Atoi: parsing \"squiggle\": invalid syntax")
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

func TestCopyDir(t *testing.T) {
	// Create a temporary directory for the repository
	srcDir, err := os.MkdirTemp("", "src")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(srcDir)

	// Create a temporary directory for the repository
	destDir, err := os.MkdirTemp("", "dest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(destDir)

	// Create a file in the source directory
	srcFile, err := os.Create(srcDir + "/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	srcFile.Close()

	// Copy the directory
	err = copyDir(srcDir, destDir)
	assert.NoError(t, err)

	// Check that the file was copied
	_, err = os.Stat(destDir + "/file.txt")
	assert.NoError(t, err)
}

// TestCopyDir_SrcDirError tests the error handling of the copyDir function when the source directory does not exist
func TestCopyDir_SrcDirError(t *testing.T) {
	// Create a temporary directory for the repository
	destDir, err := os.MkdirTemp("", "dest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(destDir)

	// Copy the directory
	err = copyDir("nonexistent", destDir)
	assert.Error(t, err)

	// Check that the error is as expected
	assert.EqualError(t, err, "error getting source directory info: stat nonexistent: no such file or directory")
}

// TestCopyDir_SrcDirIsFileError tests the error handling of the copyDir function when the source directory is a file
func TestCopyDir_SrcDirIsFileError(t *testing.T) {
	// Create a temporary directory for the repository
	srcDir, err := os.MkdirTemp("", "src")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(srcDir)

	// Create a file in the source directory
	srcFile, err := os.Create(srcDir + "/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	srcFile.Close()

	// Create a temporary directory for the repository
	destDir, err := os.MkdirTemp("", "dest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(destDir)

	// Copy the directory
	err = copyDir(srcDir+"/file.txt", destDir)
	assert.Error(t, err)

	// Check that the error is as expected
	assert.EqualError(t, err, srcDir+"/file.txt is not a directory")
}

// TestExtractKeyFromQuery tests the successful extraction of a given key from a query string
func TestExtractKeyFromQuery(t *testing.T) {
	src := "https://example.com/org/repo.git?ref=foo//bar"
	u, _ := url.Parse(src)
	subdir := "/bar"

	ref := extractKeyFromQuery(u.Query(), "ref", &subdir)
	assert.Equal(t, "foo", ref)
}

func TestExtractSubdirFromQuery(t *testing.T) {
	src := "https://example.com/org/repo.git"
	u, _ := url.Parse(src)
	subdir := ""

	ref := extractKeyFromQuery(u.Query(), "ref", &subdir)
	assert.Equal(t, "", ref)
}
