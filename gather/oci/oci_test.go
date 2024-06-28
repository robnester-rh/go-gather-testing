package oci

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
)

func getRegistryURL(src string) string {
	parts := strings.Split(src, "/")
	lastPart := parts[len(parts)-1]
	if strings.Contains(lastPart, ":") {
		return src
	}
	return src + ":latest"
}

// TestGetRegistryURL tests the getRegistryURL function.
func TestGetRegistryURL(t *testing.T) {
	testCases := []struct {
		src      string
		expected string
	}{
		{src: "docker.io/library/alpine", expected: "docker.io/library/alpine:latest"},
		{src: "docker.io/library/alpine:3.12", expected: "docker.io/library/alpine:3.12"},
		{src: "https://docker.io/library/alpine", expected: "https://docker.io/library/alpine:latest"},
		{src: "alpine", expected: "alpine:latest"},
	}

	for _, tc := range testCases {
		actual := getRegistryURL(tc.src)
		if actual != tc.expected {
			t.Errorf("Expected getRegistryURL(%s) to return %s, but got %s", tc.src, tc.expected, actual)
		}
	}
}

func TestOCIURLParse(t *testing.T) {
	testCases := []struct {
		source   string
		expected string
	}{
		{source: "docker.io/library/alpine", expected: "docker.io/library/alpine"},
		{source: "https://docker.io/library/alpine", expected: "docker.io/library/alpine"},
		{source: "alpine", expected: "alpine"},
		{source: "https://example.com/image:tag", expected: "example.com/image:tag"},
	}

	for _, tc := range testCases {
		actual := ociURLParse(tc.source)
		if actual != tc.expected {
			t.Errorf("Expected ociURLParse(%s) to return %s, but got %s", tc.source, tc.expected, actual)
		}
	}
}

// TestOCIGatherer_Gather tests the Gather function.
func TestOCIGatherer_Gather(t *testing.T) {
	ctx := context.TODO()

	testCases := []struct {
		name         string
		source       string
		destination  string
		expectedRepo string
		expectedErr  error
	}{
		{
			name:         "Valid source URI",
			source:       "quay.io/libpod/alpine",
			destination:  "/tmp/foo",
			expectedRepo: "quay.io/libpod/alpine:latest",
			expectedErr:  nil,
		},
		{
			name:         "Valid source URI with tag",
			source:       "quay.io/libpod/alpine:3.2",
			destination:  "/tmp/foo",
			expectedRepo: "quay.io/libpod/alpine:3.2",
			expectedErr:  nil,
		},
		{
			name:         "Valid source URI with HTTPS",
			source:       "https://quay.io/libpod/alpine",
			destination:  "/tmp/foo",
			expectedRepo: "https://quay.io/libpod/alpine:latest",
			expectedErr:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gatherer := &OCIGatherer{}
			_, err := gatherer.Gather(ctx, tc.source, tc.destination)

			if err != tc.expectedErr {
				t.Errorf("Expected error: %v, but got: %v", tc.expectedErr, err)
			}
		})
		t.Cleanup(func() {
			// Cleanup the destination directory
			os.RemoveAll(tc.destination)
		})
	}
}

// TestOCIGatherer_Gather_Invalid_URIs tests the Gather function with invalid source URIs.
func TestOCIGatherer_Gather_Invalid_URIs(t *testing.T) {
	ctx := context.TODO()

	testCases := []struct {
		name        string
		source      string
		destination string
		expectedErr error
	}{
		{
			name:        "Invalid source URI",
			source:      "invalid",
			destination: "/tmp/foo",
			expectedErr: fmt.Errorf("failed to parse reference: invalid reference: missing registry or repository"),
		},
		{
			name:        "Invalid source URI with tag",
			source:      "invalid:tag",
			destination: "/tmp/foo",
			expectedErr: fmt.Errorf("failed to parse reference: invalid reference: missing registry or repository"),
		},
		{
			name:        "Invalid source URI with HTTPS",
			source:      "https://invalid",
			destination: "/tmp/foo",
			expectedErr: fmt.Errorf("failed to parse reference: invalid reference: missing registry or repository"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gatherer := &OCIGatherer{}
			metadata, err := gatherer.Gather(ctx, tc.source, tc.destination)

			if err.Error() != tc.expectedErr.Error() {
				t.Errorf("Expected error: %v, but got: %v", tc.expectedErr, err)
			}

			if metadata != nil {
				t.Errorf("Expected metadata to be nil, but got: %v", metadata)
			}
		})
		t.Cleanup(func() {
			// Cleanup the destination directory
			os.RemoveAll(tc.destination)
		})
	}
}

// TestOCIGatherer_Gather_ErorrCreatingNewRepository tests the Gather function with an error creating a new repository client.
func TestOCIGatherer_Gather_ErorrCreatingNewRepository(t *testing.T) {
	testCases := []struct {
		name        string
		source      string
		destination string
		expectedErr error
	}{
		{
			name:        "Error creating new repository",
			source:      "docker.io",
			destination: "/tmp/foo",
			expectedErr: fmt.Errorf("failed to parse reference: invalid reference: missing registry or repository"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.TODO()
			gatherer := &OCIGatherer{}
			metadata, err := gatherer.Gather(ctx, tc.source, tc.destination)

			if err.Error() != tc.expectedErr.Error() {
				t.Errorf("Expected error: %v, but got: %v", tc.expectedErr, err)
			}

			if metadata != nil {
				t.Errorf("Expected metadata to be nil, but got: %v", metadata)
			}
		})
		t.Cleanup(func() {
			// Cleanup the destination directory
			os.RemoveAll(tc.destination)
		})
	}

}
