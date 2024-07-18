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

package oci

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"oras.land/oras-go/v2"

	"github.com/enterprise-contract/go-gather/metadata/oci"
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

// TestOCIGatherer_Gather_Success tests the Gather function when it's successful.
func TestOCIGatherer_Gather_Success(t *testing.T) {
	ctx := context.TODO()
	source := "example.com/org/repo"
	destination := "/tmp/foo"
	orasCopy = func(_ context.Context, _ oras.ReadOnlyTarget, _ string, _ oras.Target, _ string, _ oras.CopyOptions) (ocispec.Descriptor, error) {
		return ocispec.Descriptor{Digest: "fa93b01658e3a5a1686dc3ae55f170d8de487006fb53a28efcd12ab0710a2e5f"}, nil
	}

	t.Run("Gather", func(t *testing.T) {
		gatherer := &OCIGatherer{}
		m, err := gatherer.Gather(ctx, source, destination)
		if err != nil {
			t.Errorf("Expected error to be nil, but got: %v", err)
		}
		assert.Equal(t, "fa93b01658e3a5a1686dc3ae55f170d8de487006fb53a28efcd12ab0710a2e5f", m.(*oci.OCIMetadata).Digest, "Digest should be equal, expected: %s, got: %s", "fa93b01658e3a5a1686dc3ae55f170d8de487006fb53a28efcd12ab0710a2e5f", m.(*oci.OCIMetadata).Digest)
	})
	t.Cleanup(func() {
		// Cleanup the destination directory
		os.RemoveAll(destination)
	})
}

// TestOCIGatherer_Gather_Failure tests the Gather function when it fails.
func TestOCIGatherer_Gather_Failure(t *testing.T) {
	ctx := context.TODO()
	source := "example.com/org/repo"
	destination := "/tmp/foo"
	expectedError := fmt.Errorf("error")
	orasCopy = func(_ context.Context, _ oras.ReadOnlyTarget, _ string, _ oras.Target, _ string, _ oras.CopyOptions) (ocispec.Descriptor, error) {
		return ocispec.Descriptor{}, expectedError
	}

	t.Run("Gather", func(t *testing.T) {
		gatherer := &OCIGatherer{}
		_, err := gatherer.Gather(ctx, source, destination)
		assert.Equal(t, fmt.Errorf("pulling policy: %w", expectedError), err, "Error should be equal, expected: %s, got: %s", err)
	})
	t.Cleanup(func() {
		// Cleanup the destination directory
		os.RemoveAll(destination)
	})
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
