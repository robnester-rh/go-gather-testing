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
	"fmt"
	"os"
	"testing"
)

// TestFileSaver_Save tests the Save method of the FileSaver type.
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

// TestFileSaver_UrlParseError tests the Save method of the FileSaver type when the destination URI is invalid.
func TestFileSaver_UrlParseError(t *testing.T) {
	// Create a FileSaver instance
	fs := &FileSaver{}

	// Call the Save method with an invalid destination URI
	err := fs.Save(context.Background(), nil, ":")
	if err == nil {
		t.Error("expected an error, but got nil")
	}

	expectedErrorMessage := "parse \":\": missing protocol scheme"
	if err.Error() != expectedErrorMessage {
		t.Errorf("unexpected error message: got %s, want %s", err.Error(), expectedErrorMessage)
	}
}

// TestFileSaver_MkdirAllError tests the Save method of the FileSaver type when the destination directory cannot be created.
func TestFileSaver_MkdirAllError(t *testing.T) {
	// Create a FileSaver instance
	fs := &FileSaver{}
	tempDir, err := os.MkdirTemp("", "root")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	os.Chmod(tempDir, 0000)

	destination := "file://" + tempDir + "/foo/test.txt"
	// Call the Save method with a destination path that cannot be created
	err = fs.Save(context.Background(), nil, destination)
	if err == nil {
		t.Error("expected an error, but got nil")
	}

	expectedErrorMessage := fmt.Sprintf("failed to create destination directory: mkdir %s/foo: permission denied", tempDir)
	if err.Error() != expectedErrorMessage {
		t.Errorf("unexpected error message: got %s, want %s", err.Error(), expectedErrorMessage)
	}
}

// TestFileSaver_OsCreateError tests the Save method of the FileSaver type when the destination file cannot be created.
func TestFileSaver_OsCreateError(t *testing.T) {
	// Create a FileSaver instance
	fs := &FileSaver{}

	// Call the Save method with a destination path that cannot be created
	err := fs.Save(context.Background(), nil, "/root/test.txt")
	if err == nil {
		t.Error("expected an error, but got nil")
	}

	expectedErrorMessage := "open /root/test.txt: permission denied"
	if err.Error() != expectedErrorMessage {
		t.Errorf("unexpected error message: got %s, want %s", err.Error(), expectedErrorMessage)
	}
}