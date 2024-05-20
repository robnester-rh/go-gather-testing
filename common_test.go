package gogather

import (
	"fmt"
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
	GetHomeDir = func() (string, error) {
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
	GetHomeDir = func() (string, error) {
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
		input string
		expected URIType
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
