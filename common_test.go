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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestURITypeString tests the String method of the URIType type.
func TestURITypeString(t *testing.T) {
	testCases := []struct {
		input    URIType
		expected string
	}{
		{input: GitURI, expected: "GitURI"},
		{input: HTTPURI, expected: "HTTPURI"},
		{input: FileURI, expected: "FileURI"},
		{input: Unknown, expected: "Unknown"},
	}

	for _, tc := range testCases {
		actual := tc.input.String()
		if actual != tc.expected {
			t.Errorf("Expected %s.String() to return %s, but got %s", tc.input, tc.expected, actual)
		}
	}
}

// TestExpandTilde tests the ExpandTilde function.
func TestExpandTilde(t *testing.T) {
	getHomeDir = func() (string, error) {
		return "/home/user", nil
	}

	testCases := []struct {
		path     string
		expected string
	}{
		{path: "~/Documents/file.txt", expected: "/home/user/Documents/file.txt"},
		{path: "/var/www/html/index.html", expected: "/var/www/html/index.html"},
		{path: "file::/home/user/file.txt", expected: "file::/home/user/file.txt"},
	}

	for _, tc := range testCases {
		actual := ExpandTilde(tc.path)
		if actual != tc.expected {
			t.Errorf("Expected ExpandTilde(%s) to return %s, but got %s", tc.path, tc.expected, actual)
		}
	}
}

// TestExpandTilde_OsUserHomeDirError tests the ExpandTilde function when os.UserHomeDir returns an error.
func TestExpandTilde_OsUserHomeDirError(t *testing.T) {
	// Mock os.UserHomeDir to return an error
	getHomeDir = func() (string, error) {
		return "", fmt.Errorf("mock error")
	}

	path := "~/Documents/file.txt"
	actual := ExpandTilde(path)
	if actual != path {
		t.Errorf("Expected ExpandTilde(%s) to return %s, but got %s", path, path, actual)
	}
}

// TestClassifyURI tests the ClassifyURI function.
func TestClassifyURI(t *testing.T) {
	testCases := []struct {
		input    string
		expected URIType
	}{
		{input: "git::git@github.com:user/repo.git", expected: GitURI},
		{input: "git@github.com:user/repo.git", expected: GitURI},
		{input: "http::https://github.com/user/repo.git", expected: HTTPURI},
		{input: "file::/home/user/file.txt", expected: FileURI},
		{input: "file:///home/user/file.txt", expected: FileURI},
		{input: "/home/user/file.git", expected: GitURI},
		{input: "https://example.com", expected: HTTPURI},
		{input: "ftpexamplecom", expected: Unknown},
		{input: "github.com/user/repo.git", expected: GitURI},
		{input: "gitlab.com/user/repo.git", expected: GitURI},
	}

	for _, tc := range testCases {
		actual, err := ClassifyURI(tc.input)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if actual != tc.expected {
			t.Errorf("Expected ClassifyURI(%s) to return %s, but got %s", tc.input, tc.expected, actual)
		}
	}
}

func TestClassifyURI_errors(t *testing.T) {
	testCases := []struct {
		input         string
		expected      URIType
		ExpectedError error
	}{
		{input: "ftp://foo.txt", expected: Unknown, ExpectedError: nil},
		{input: ".foo.txt", expected: Unknown, ExpectedError: nil},
	}

	for _, tc := range testCases {
		_, err := ClassifyURI(tc.input)
		if err == nil {
			t.Errorf("Expected an error, but got nil")
		}
	}
}

// TestValidateFileDestination tests the ValidateFileDestination function.
func TestValidateFileDestination(t *testing.T) {
	testCases := []struct {
		destination string
	}{
		{destination: "/path/to/file.txt"},
		{destination: "/path/to/directory/"},
	}

	for _, tc := range testCases {
		err := ValidateFileDestination(tc.destination)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

// TestValidateFileDestination_errors tests the ValidateFileDestination function with invalid destinations.
func TestValidateFileDestination_errors(t *testing.T) {
	dir, _ := os.MkdirTemp("", "path")
	err := os.WriteFile(filepath.Join(dir, "file.text"), []byte("test"), 0600)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer os.RemoveAll(dir)

	testCases := []struct {
		destination string
		errExpected bool
		expectedErr error
	}{
		{
			destination: filepath.Join(dir, "file.text"),
			errExpected: true,
			expectedErr: errors.New("destination is a file"),
		},
		{
			destination: filepath.Join(dir, "file2.text"),
			errExpected: false,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		err := ValidateFileDestination(tc.destination)
		if tc.errExpected && err == nil {
			t.Errorf("Expected an error: %s,\n but got nil", tc.expectedErr)
		}
		if !tc.errExpected && err != nil {
			t.Errorf("Expected no error, but got: %s", err)
		}
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
}
