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

// Package git provides methods for gathering git repositories.
// This package implements the Gatherer interface and provides methods for cloning git repositories,
// retrieving commit metadata, and authenticating SSH connections.
package git

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	gitUrls "github.com/whilp/git-urls"

	gogather "github.com/enterprise-contract/go-gather"
	"github.com/enterprise-contract/go-gather/metadata"
	gitMetadata "github.com/enterprise-contract/go-gather/metadata/git"
)

// GitGatherer is a struct that implements the Gatherer interface
// and provides methods for gathering git repositories.
type GitGatherer struct {
	// Authenticator is an SSHAuthenticator that provides authentication for SSH connections.
	Authenticator SSHAuthenticator
}

// SSHAuthenticator represents an interface for authenticating SSH connections.
type SSHAuthenticator interface {
	// NewSSHAgentAuth returns a new SSH agent authentication method for the given user.
	// It returns an instance of transport.AuthMethod and an error if any.
	NewSSHAgentAuth(user string) (transport.AuthMethod, error)
}

// RealSSHAuthenticator represents an implementation of the SSHAuthenticator interface.
type RealSSHAuthenticator struct{}

// NewSSHAgentAuth returns an AuthMethod that uses the SSH agent for authentication.
// It uses the specified user as the username for authentication.
func (r *RealSSHAuthenticator) NewSSHAgentAuth(user string) (transport.AuthMethod, error) {
	return ssh.NewSSHAgentAuth(user)
}

// Gather clones a Git repository from the given source URI into the specified destination directory,
// and returns the metadata of the cloned repository.
func (g *GitGatherer) Gather(ctx context.Context, source, destination string) (metadata.Metadata, error) {
	src, ref, subdir, depth, err := processUrl(source)
	if err != nil {
		return nil, fmt.Errorf("failed to process URL: %w", err)
	}

	cloneOpts := &git.CloneOptions{
		URL: src,
	}

	if ref != "" {
		cloneOpts.ReferenceName = plumbing.ReferenceName("refs/heads/" + ref)
	}

	if depth != "" {
		depth, err := strconv.Atoi(depth)
		if err != nil {
			return nil, fmt.Errorf("failed to parse depth: %w", err)
		}
		cloneOpts.Depth = depth
	}

	// If we don't have a subdir, clone the repository and return the metadata
	if subdir == "" {
		r, err := git.PlainClone(destination, false, cloneOpts)
		if err != nil {
			return nil, fmt.Errorf("error cloning repository: %w", err)
		}

		// Get the commit history
		commits, err := r.CommitObjects()
		if err != nil {
			return nil, fmt.Errorf("error getting commit history: %w", err)
		}

		// Safely accumulate commits into the metadata structure
		m := &gitMetadata.GitMetadata{}
		err = commits.ForEach(func(c *object.Commit) error {
			m.Commits = append(m.Commits, *c)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error accumulating commits: %w", err)
		}

		return m, nil

	}

	// If we have a subdir, clone the repository and copy the subdir to the destination
	return cloneRepositoryPath(ctx, subdir, destination, cloneOpts)
}

// cloneRepositoryPath clones a git repository, copies the specified subdirectory to the destination, and returns the metadata.
func cloneRepositoryPath(ctx context.Context, path, destination string, cloneOpts *git.CloneOptions) (metadata.Metadata, error) {
	// create a temporary directory to clone the repository into
	tmpDir, err := os.MkdirTemp("", "git-repo-")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Clone the repository into the temporary directory
	r, err := git.PlainCloneContext(ctx, tmpDir, false, cloneOpts)
	if err != nil {
		return nil, fmt.Errorf("error cloning repository: %w", err)
	}

	// Get the worktree
	w, err := r.Worktree()
	if err != nil {
		return nil, fmt.Errorf("error getting worktree: %w", err)
	}

	// Check if the path exists in the repository
	_, err = w.Filesystem.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("path %s does not exist in the repository", path)
	}

	path = filepath.Join(tmpDir, path)

	err = copyDir(path, destination)
	if err != nil {
		return nil, fmt.Errorf("error copying directory: %w", err)
	}

	// Get the commit history
	commits, err := r.CommitObjects()
	if err != nil {
		return nil, fmt.Errorf("error getting commit history: %w", err)
	}

	// Safely accumulate commits into the metadata structure
	m := &gitMetadata.GitMetadata{}
	err = commits.ForEach(func(c *object.Commit) error {
		m.Commits = append(m.Commits, *c)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error accumulating commits: %w", err)
	}

	return m, nil
}

// copyDir copies the contents of the src directory to dst directory
func copyDir(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("error getting source directory info: %w", err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", src)
	}

	_, err = os.Stat(dst)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dst, srcInfo.Mode())
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

// extractSubdirFromQuery extracts the value of the key from the query parameters and extracts a subdir, if present.
func extractSubdirFromQuery(q url.Values, key string, subdir *string) string {
	value := q.Get(key)
	if strings.Contains(value, "//") {
		parts := strings.SplitN(value, "//", 2)
		*subdir = parts[1]
		q.Del(key)
		return parts[0]
	}
	q.Del(key)
	return value
}

// getGitCloneOptions returns the clone options for the git repository.
func getCloneOptions(source string, auth SSHAuthenticator) (*git.CloneOptions, error) {
	src, err := gitUrls.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source URL: %w", err)
	}

	cloneOpts := &git.CloneOptions{
		URL:   src.String(),
		Depth: 1,
	}

	if src.Scheme == "git" {
		cloneOpts.URL = strings.Replace(source, "git::", "", 1)
	}

	if src.Scheme == "ssh" {
		authMethod, err := auth.NewSSHAgentAuth("git")
		if err != nil {
			return nil, fmt.Errorf("failed to create SSH auth method: %w", err)
		}
		cloneOpts.Auth = authMethod
	}

	return cloneOpts, nil
}

// processUrl processes the raw URL and returns the source URL, ref, subdir, and depth.
func processUrl(rawURL string) (src, ref, subdir, depth string, err error) {
	// Check if the URL is a git URL and if it is not a SSH URL, convert it to HTTPS
	t, err := gogather.ClassifyURI(rawURL)
	if err != nil {
		return src, ref, subdir, depth, fmt.Errorf("failed to classify URI: %w", err)
	}
	if t == gogather.GitURI && !strings.Contains(rawURL, "git@") && !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	// Check if the rawURL contains "::" and split it to get the actual URL if it does
	if strings.Contains(rawURL, "::") {
		rawURL = strings.Split(rawURL, "::")[1]
	}

	// Parse the raw URL with the gitUrls package. This will format the URL correctly
	parsedURL, err := gitUrls.Parse(rawURL)
	if err != nil {
		return src, ref, subdir, depth, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Parse the URL again with the url package to extract the query parameters, etc.
	u, err := url.Parse(parsedURL.String())
	if err != nil {
		return src, ref, subdir, depth, fmt.Errorf("failed to reparse URL: %w", err)
	}

	// Extract the ref, subdir, and depth from the query parameters
	q := u.Query()
	ref = extractSubdirFromQuery(q, "ref", &subdir)
	depth = extractSubdirFromQuery(q, "depth", &subdir)
	u.RawQuery = q.Encode()

	// If the path contains "//", split it to get the actual path and subdir
	if strings.Contains(u.Path, "//") {
		parts := strings.SplitN(u.Path, "//", 2)
		u.Path = parts[0]
		subdir = parts[1]
	}

	// If the path does not end with ".git", append it
	if !strings.HasSuffix(u.Path, ".git") {
		u.Path += ".git"
	}

	// Return the URL, ref, subdir, and depth
	return u.String(), ref, subdir, depth, nil
}
