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

package gogather

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// URLType is an enum for URL types
type URIType int

const (
	GitURI URIType = iota
	HTTPURI
	FileURI
	OCIURI
	Unknown
)

var getHomeDir = os.UserHomeDir

// String returns the string representation of the URLType
func (t URIType) String() string {
	return [...]string{"GitURI", "HTTPURI", "FileURI", "OCIURI", "Unknown"}[t]
}

// ExpandTilde expands a leading tilde in the file path to the user's home directory
func ExpandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := getHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

// ClassifyURI classifies the input string as a Git URI, HTTP(S) URI, or file path
func ClassifyURI(input string) (URIType, error) {
	// Check for special prefixes first
	if strings.HasPrefix(input, "git::") {
		return GitURI, nil
	}
	if strings.HasPrefix(input, "file::") {
		return FileURI, nil
	}
	if strings.HasPrefix(input, "http::") {
		return HTTPURI, nil
	}

	if strings.HasPrefix(input, "oci::") {
		return OCIURI, nil
	}

	if strings.HasPrefix(input, "github.com") || strings.HasPrefix(input, "gitlab.com") {
		return GitURI, nil
	}

	// Regular expression for Git URIs
	gitURIPattern := regexp.MustCompile(`^(git@[\w\.\-]+:[\w\.\-]+/[\w\.\-]+(\.git)?|https?://[\w\.\-]+/[\w\.\-]+/[\w\.\-]+(\.git)?|git://[\w\.\-]+/[\w\.\-]+/[\w\.\-]+(\.git)?|[\w\.\-]+/[\w\.\-]+/[\w\.\-]+//.*|file://.*\.git|[\w\.\-]+/[\w\.\-]+(\.git)?)$`)
	// Regular expression for HTTP URIs (with or without protocol)
	httpURIPattern := regexp.MustCompile(`^((http://|https://)[\w\-]+(\.[\w\-]+)+.*)$`)
	// Regular expression for file paths
	filePathPattern := regexp.MustCompile(`^(\./|\../|/|[a-zA-Z]:\\|~\/|file://).*`)
	// Regular expression for OCI URIs
	ociURIPattern := regexp.MustCompile(`^((oci://)[\w\-]+(\.[\w\-]+)+.*)$`)
	// Regular expressions for known OCI registries

	// Check if the input matches the file path pattern first
	if filePathPattern.MatchString(input) {
		// Expand the tilde in the file path if it exists
		input = ExpandTilde(input)
		// Check if the input ends with ".git" to classify as GitURI
		if strings.HasSuffix(input, ".git") {
			return GitURI, nil
		}
		return FileURI, nil
	}

	// Check if the input matches the Git URI pattern
	if gitURIPattern.MatchString(input) {
		return GitURI, nil
	}

	// Check if the input matches the HTTP URI pattern
	if httpURIPattern.MatchString(input) {
		// Parse the input as a URI
		parsedURI, err := url.Parse(input)
		if err == nil && (parsedURI.Scheme == "http" || parsedURI.Scheme == "https") {
			return HTTPURI, nil
		}
	}

	// Check if the input matches the OCI URI pattern
	if ociURIPattern.MatchString(input) {
		return OCIURI, nil
	}

	// Check if the input matches any known OCI registry
	isOCI := containsOCIRegistry(input)
	if isOCI {
		return OCIURI, nil
	}

	// Check for unsupported schemes
	parsedURI, err := url.Parse(input)
	if err == nil && parsedURI.Scheme != "" && parsedURI.Scheme != "http" && parsedURI.Scheme != "https" {
		return Unknown, fmt.Errorf("unsupported source protocol: %s", parsedURI.Scheme)
	}

	// Check if the input contains a dot but lacks a valid scheme
	if strings.Contains(input, ".") {
		return Unknown, fmt.Errorf("got %s. HTTP(S) URIs require a scheme (http:// or https://)", input)
	}

	return Unknown, nil
}

// ValidateFileDestination validates the destination path for saving files
func ValidateFileDestination(destination string) error {
	// Expand the tilde in the file path if it exists
	destination = ExpandTilde(destination)
	// Check if the destination file exists.
	_, err := os.Stat(destination)
	if err == nil {
		return fmt.Errorf("destination file already exists: %s", destination)
	}
	if os.IsNotExist(err) {
		return nil
	}
	return nil
}

// containsOCIRegistry checks if the input string contains a known OCI registry
func containsOCIRegistry(src string) bool {
	matchRegistries := []*regexp.Regexp{
		regexp.MustCompile("azurecr.io"),
		regexp.MustCompile("gcr.io"),
		regexp.MustCompile("registry.gitlab.com"),
		regexp.MustCompile("pkg.dev"),
		regexp.MustCompile("[0-9]{12}.dkr.ecr.[a-z0-9-]*.amazonaws.com"),
		regexp.MustCompile("^quay.io"),
		regexp.MustCompile(`(?:::1|127\.0\.0\.1|(?i:localhost)):\d{1,5}`), // localhost OCI registry
	}

	for _, matchRegistry := range matchRegistries {
		if matchRegistry.MatchString(src) {
			return true
		}
	}
	return false
}
