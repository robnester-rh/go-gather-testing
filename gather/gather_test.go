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

package gather

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	gogather "github.com/enterprise-contract/go-gather"
	"github.com/enterprise-contract/go-gather/metadata"
	"github.com/enterprise-contract/go-gather/metadata/git"
)

func TestGather(t *testing.T) {
	ctx := context.Background()
	t.Run("SourceParseError", func(t *testing.T) {
		source := ":"
		destination := "/tmp/foo"

		_, err := Gather(ctx, source, destination)
		if err == nil {
			t.Error("expected an error, but got nil")
		}

		expectedErrorMessage := "unsupported source protocol: Unknown"
		if err.Error() != expectedErrorMessage {
			t.Errorf("expected error message: %s, but got: %s", expectedErrorMessage, err.Error())
		}
		t.Cleanup(func() {
			os.RemoveAll(destination)
		})
	})

	t.Run("UnsupportedProtocol", func(t *testing.T) {
		source := "ftp://example.com/file.txt"
		destination := "/tmp/foo"
		defer os.RemoveAll(destination)

		_, err := Gather(ctx, source, destination)
		if err == nil {
			t.Error("expected an error, but got nil")
		}

		expectedErrorMessage := "failed to classify source URI: unsupported source protocol: ftp"
		if err.Error() != expectedErrorMessage {
			t.Errorf("expected error message: %s, but got: %s", expectedErrorMessage, err.Error())
		}
		t.Cleanup(func() {
			os.RemoveAll(destination)
		})
	})

	t.Run("SupportedProtocol_git", func(t *testing.T) {
		source := "git::https://github.com/git-fixtures/basic.git"
		destination := "/tmp/foo"
		defer os.RemoveAll(destination)

		_, err := Gather(ctx, source, destination)
		if err != nil {
			t.Errorf("expected no error, but got: %s", err.Error())
		}
		t.Cleanup(func() {
			os.RemoveAll(destination)
		})
	})

	t.Run("SupportedProtocol_file", func(t *testing.T) {
		source := "file:///tmp/foo.txt"
		destination := "file:///tmp/bar.txt"
		src, _ := url.Parse(source)
		dst, _ := url.Parse(destination)
		_ = os.WriteFile(src.Path, []byte("hello world"), 0600)
		defer os.RemoveAll(src.Path)
		defer os.RemoveAll(dst.Path)

		_, err := Gather(ctx, src.Path, destination)
		if err != nil {
			t.Errorf("expected no error, but got: %s", err.Error())
		}
		t.Cleanup(func() {
			os.RemoveAll(destination)
		})
	})

	t.Run("CustomGatherer", func(t *testing.T) {
		source := "custom_source"
		destination := "custom_destination"

		gatherer := &mockGatherer{}
		_, err := gatherer.Gather(ctx, source, destination)
		if err != nil {
			t.Errorf("expected no error, but got: %s", err.Error())
		}
		t.Cleanup(func() {
			os.RemoveAll(destination)
		})
	})
}

type mockGatherer struct{}

func (m *mockGatherer) Gather(ctx context.Context, source, destination string) (metadata.Metadata, error) {
	// Mock implementation
	return &git.GitMetadata{}, nil
}
func TestExpandTilde(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	t.Run("NoTilde", func(t *testing.T) {
		path := "/path/to/file"
		expandedPath := gogather.ExpandTilde(path)
		if expandedPath != path {
			t.Errorf("expected expanded path: %s, but got: %s", path, expandedPath)
		}
	})

	t.Run("WithTilde", func(t *testing.T) {
		path := "~/path/to/file"
		expectedPath := filepath.Join(homeDir, "path/to/file")
		expandedPath := gogather.ExpandTilde(path)
		if expandedPath != expectedPath {
			t.Errorf("expected expanded path: %s, but got: %s", expectedPath, expandedPath)
		}
	})

	t.Run("WithTildeSlash", func(t *testing.T) {
		path := "~/path/to/file/"
		expectedPath := filepath.Join(homeDir, "path/to/file/")
		expandedPath := gogather.ExpandTilde(path)
		if expandedPath != expectedPath {
			t.Errorf("expected expanded path: %s, but got: %s", expectedPath, expandedPath)
		}
	})
}
