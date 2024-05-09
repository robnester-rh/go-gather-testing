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

package file

import (
	"bytes"
	"context"
	"os"
	"testing"
)

func TestFileSaver_Save(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Create a FileSaver instance
	fs := &FileSaver{}

	// Prepare test data
	testData := []byte("test data")

	// Call the Save method
	err = fs.Save(context.Background(), bytes.NewReader(testData), tempFile.Name())
	if err != nil {
		t.Fatalf("failed to save file: %v", err)
	}

	// Read the saved file
	savedData, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	// Assert the saved data
	if !bytes.Equal(savedData, testData) {
		t.Errorf("unexpected saved data: got %s, want %s", savedData, testData)
	}
}
