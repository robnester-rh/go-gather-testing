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

// Package file provides functionality for saving data to local filesystem paths.
//
// This package contains the FileSaver type, which implements the Saver interface
// for file destinations. It allows saving data from an io.Reader to a specified
// destination path on the local filesystem.
//
// Example usage:
//   fs := &file.FileSaver{}
//   err := fs.Save(context.Background(), data, "/path/to/destination/file.txt")
//   if err != nil {
//     log.Fatal(err)
//   }
package file

import (
	"context"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

// FileSaver handles saving data to local filesystem paths.
type FileSaver struct{}

// Save implements the Saver interface for file destinations.
func (fs *FileSaver) Save(ctx context.Context, data io.Reader, destination string) error {

	dst, err := url.Parse(destination)
	if err != nil {
		return err
	}

	// Ensure the destination directory exists.
	if err := os.MkdirAll(filepath.Dir(dst.Path), 0755); err != nil {
		return err
	}

	// Create the destination file.
	f, err := os.Create(dst.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the data to the file.
	_, err = io.Copy(f, data)
	return err
}
