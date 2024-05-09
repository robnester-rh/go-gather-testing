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
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	gitUrls "github.com/whilp/git-urls"

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

// Gather clones a Git repository from the given source URL into the specified destination directory,
// and returns the metadata of the cloned repository.
func (g *GitGatherer) Gather(ctx context.Context, source, destination string) (metadata.Metadata, error) {
	// Get our clone options
	cloneOpts, err := getCloneOptions(source, g.Authenticator)
	if err != nil {
		return nil, err
	}

	// We assume that destination is a unique directory and clone into that.
	r, err := git.PlainClone(destination, false, cloneOpts)
	if err != nil {
		return nil, err
	}

	commits, err := r.CommitObjects()
	if err != nil {
		return nil, err
	}

	// Safely accumulate commits into the metadata structure
	m := &gitMetadata.GitMetadata{}
	err = commits.ForEach(func(c *object.Commit) error {
		m.Commits = append(m.Commits, *c)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

// getGitCloneOptions returns the clone options for the git repository.
func getCloneOptions(source string, auth SSHAuthenticator) (*git.CloneOptions, error) {
	src, err := gitUrls.Parse(source)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		cloneOpts.Auth = authMethod
	}

	return cloneOpts, nil
}
