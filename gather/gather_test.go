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
	"os"
	"testing"

	"github.com/enterprise-contract/go-gather/metadata"
	"github.com/enterprise-contract/go-gather/metadata/git"
)

func TestGather(t *testing.T) {
	ctx := context.Background()
	t.Run("SourceParseError", func(t *testing.T) {
		source := ":"
		destination := "/path/to/destination"

		_, err := Gather(ctx, source, destination)
		if err == nil {
			t.Error("expected an error, but got nil")
		}

		expectedErrorMessage := "parse \":\": missing protocol scheme"
		if err.Error() != expectedErrorMessage {
			t.Errorf("expected error message: %s, but got: %s", expectedErrorMessage, err.Error())
		}
	})

	t.Run("UnsupportedProtocol", func(t *testing.T) {
		source := "ftp://example.com/file.txt"
		destination := "/path/to/destination"

		_, err := Gather(ctx, source, destination)
		if err == nil {
			t.Error("expected an error, but got nil")
		}

		expectedErrorMessage := "unsupported source protocol: ftp"
		if err.Error() != expectedErrorMessage {
			t.Errorf("expected error message: %s, but got: %s", expectedErrorMessage, err.Error())
		}
	})

	t.Run("SupportedProtocol_git", func(t *testing.T) {
		source := "git::https://github.com/git-fixtures/basic.git"
		destination := "/tmp/path/to/destination"
		defer os.RemoveAll(destination)

		// Add your test logic here
		// BEGIN: SupportedProtocolTest
		_, err := Gather(ctx, source, destination)
		if err != nil {
			t.Errorf("expected no error, but got: %s", err.Error())
		}
		// END: SupportedProtocolTest
		t.Cleanup(func() {
			os.RemoveAll(destination)
		})
	})

	t.Run("SupportedProtocol_file", func(t *testing.T) {
		source := "file:///tmp/foo.txt"
		destination := "file:///tmp/bar.txt"
		defer os.RemoveAll(destination)
		_ = os.WriteFile(source, []byte("hello world"), 0600)
		defer os.RemoveAll(source)

		// Add your test logic here
		// BEGIN: SupportedProtocolTest
		_, err := Gather(ctx, source, destination)
		if err != nil {
			t.Errorf("expected no error, but got: %s", err.Error())
		}
		// END: SupportedProtocolTest
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
	})
}

type mockGatherer struct{}

func (m *mockGatherer) Gather(ctx context.Context, source, destination string) (metadata.Metadata, error) {
	// Mock implementation
	return &git.GitMetadata{}, nil
}
