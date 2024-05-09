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

package saver

import (
	"fmt"
	"testing"

	"github.com/enterprise-contract/go-gather/saver/file"
)

func TestNewSaver(t *testing.T) {
	// Test case 1: protocol is "file"
	protocol := "file"
	saver, err := NewSaver(protocol)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	_, ok := saver.(*file.FileSaver)
	if !ok {
		t.Errorf("unexpected saver type: got %T, want *file.FileSaver", saver)
	}

	// Test case 2: unsupported protocol
	protocol = "unsupported"
	_, err = NewSaver(protocol)
	expectedErr := fmt.Errorf("unsupported protocol: %s", protocol)
	if err == nil {
		t.Error("expected an error, but got nil")
	} else if err.Error() != expectedErr.Error() {
		t.Errorf("unexpected error: got %v, want %v", err, expectedErr)
	}
}
