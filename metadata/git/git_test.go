// Copyright The Enterprise Contract Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"fmt"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"

	"github.com/enterprise-contract/go-gather/metadata"
)

func TestGitMetadata_Get(t *testing.T) {
	metadata := GitMetadata{
		LatestCommit: plumbing.ComputeHash(plumbing.AnyObject, []byte("hash1")).String(),
	}

	expectedResult := map[string]any{
		"latest_commit": metadata.LatestCommit,
	}
	result := metadata.Get()

	assert.Equal(t, expectedResult, result, fmt.Sprintf("expected: %v, got: %v", expectedResult, result))
}

// TestGetLatestCommit tests the GetLatestCommit method.
func TestGitMetadata_GetLatestCommit(t *testing.T) {
	metadata := GitMetadata{
		LatestCommit: plumbing.ComputeHash(plumbing.AnyObject, []byte("hash1")).String(),
	}

	expectedResult := metadata.LatestCommit
	result := metadata.GetLatestCommit()

	assert.Equal(t, expectedResult, result, fmt.Sprintf("expected: %v, got: %v", expectedResult, result))
}

func TestGetPinnedUrl(t *testing.T) {
	goodMetadata := GitMetadata{
		LatestCommit: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	}
	badMetadata := GitMetadata{}

	tests := []struct {
		name          string
		url           string
		expectedURL   string
		expectError   bool
		expectedError string
		metadata      GitMetadata
	}{
		{
			name:        "valid URL with git:// scheme",
			url:         "git://example.com/org/repo.git",
			expectedURL: "git::example.com/org/repo.git?ref=e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expectError: false,
			metadata:    goodMetadata,
		},
		{
			name:        "valid URL with git:: scheme",
			url:         "git::example.com/org/repo.git",
			expectedURL: "git::example.com/org/repo.git?ref=e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expectError: false,
			metadata:    goodMetadata,
		},
		{
			name:        "valid URL with git:: scheme",
			url:         "git::git://example.com/org/repo.git",
			expectedURL: "git::example.com/org/repo.git?ref=e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expectError: false,
			metadata:    goodMetadata,
		},
		{
			name:        "valid URL with https:// scheme",
			url:         "https://example.com/org/repo.git",
			expectedURL: "git::example.com/org/repo.git?ref=e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expectError: false,
			metadata:    goodMetadata,
		},
		{
			name:        "valid URL without .git extension",
			url:         "git://example.com/org/repo",
			expectedURL: "git::example.com/org/repo?ref=e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expectError: false,
			metadata:    goodMetadata,
		},
		{
			name:        "valid git@ URL",
			url:         "git@example.com:org/repo.git",
			expectedURL: "git::example.com/org/repo.git?ref=e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expectError: false,
			metadata:    goodMetadata,
		},
		{
			name:          "invalid URL",
			url:           "",
			expectedURL:   "",
			expectError:   true,
			expectedError: "empty URL",
			metadata:      goodMetadata,
		},
		{
			name:          "valid URL with empty metadata",
			url:           "git://example.com/org/repo.git",
			expectedURL:   "git::example.com/org/repo.git?ref=e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expectError:   true,
			expectedError: "latest commit not set",
			metadata:      badMetadata,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.metadata.GetPinnedURL(tt.url)
			if tt.expectError && err != nil {
				assert.Equal(t, err.Error(), tt.expectedError, fmt.Sprintf("GetPinnedURL() error = %v, expectedError %v", err, tt.expectedError))
				return
			}
			assert.Equal(t, result, tt.expectedURL, fmt.Sprintf("GetPinnedURL() gotURL = %v, expectedURL %v", result, tt.expectedURL))
		})
	}
}

func TestGetPinnedURL(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		metadata metadata.Metadata
		expected string
		hasError bool
	}{
		// Git Metadata Tests
		{
			name: "Git URL with git:: prefix and ref",
			url:  "git::https://test-url.git?ref=abc1234",
			metadata: &GitMetadata{
				LatestCommit: "def456",
			},
			expected: "git::test-url.git?ref=def456",
			hasError: false,
		},
		{
			name: "Git URL without git:: prefix",
			url:  "https://test-url.git?ref=abc1234",
			metadata: &GitMetadata{
				LatestCommit: "def456",
			},
			expected: "git::test-url.git?ref=def456",
			hasError: false,
		},
		{
			name: "Git URL without git:: prefix and @git",
			url:  "git@test-url.com:org/repo.git?ref=abc1234",
			metadata: &GitMetadata{
				LatestCommit: "def456",
			},
			expected: "git::test-url.com/org/repo.git?ref=def456",
			hasError: false,
		},
		{
			name: "Git URL with git:: prefix and path suffix",
			url:  "git::https://test-url.git//path/to/file?ref=abc1234",
			metadata: &GitMetadata{
				LatestCommit: "ghi789",
			},
			expected: "git::test-url.git//path/to/file?ref=ghi789",
			hasError: false,
		},
		{
			name: "Git URL with git:: prefix, path suffix, and existing SHA (should ignore SHA)",
			url:  "git::https://test-url.git//path/to/file?ref=abc1234",
			metadata: &GitMetadata{
				LatestCommit: "ghi789",
			},
			expected: "git::test-url.git//path/to/file?ref=ghi789",
			hasError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // Run tests in parallel where possible

			got, err := tc.metadata.GetPinnedURL(tc.url)
			if (err != nil) != tc.hasError {
				t.Errorf("GetPinnedURL() \nerror = %v, \nexpected error = %v", err, tc.hasError)
				t.Fatalf("GetPinnedURL() \nerror = %v, \nexpected error = %v", err, tc.hasError)
			}
			if got != tc.expected {
				t.Errorf("GetPinnedURL() = %q\ninput = %q\nexpected = %q\ngot = %q", got, tc.url, tc.expected, got)
			}
		})
	}
}
