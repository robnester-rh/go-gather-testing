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

// Package file provides functionality for copying files and directories.
// It includes a FileGatherer struct that implements the Gatherer interface
// and provides methods for gathering files and directories.
package file

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/enterprise-contract/go-gather/expander"
	"github.com/enterprise-contract/go-gather/metadata"
	"github.com/enterprise-contract/go-gather/metadata/file"
	"github.com/enterprise-contract/go-gather/saver"
)

// FileGatherer is a struct that implements the Gatherer interface
// and provides methods for gathering files and directories.
type FileGatherer struct{}

// Gather copies a file or directory from the source path to the destination path.
// It returns the metadata of the gathered file or directory and any error encountered.
func (f *FileGatherer) Gather(ctx context.Context, source, destination string) (metadata.Metadata, error) {
	// Parse the source URI
	src, err := url.Parse(source)

	if err != nil {
		return nil, fmt.Errorf("failed to parse source URI: %w", err)
	}

	// Determine if we have a file or directory
	sourceKind, err := os.Stat(src.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to determine source kind: %w", err)
	}

	// Determine if we have a tar file as the src. If so, we need to untar it.
	if strings.HasSuffix(src.Path, ".tar") {
		dst, err := url.Parse(destination)
		if err != nil {
			return nil, fmt.Errorf("failed to parse destination URI: %w", err)
		}

		t := &expander.TarExpander{
			FilesLimit:    0,
			FileSizeLimit: 0,
		}

		err = t.Expand(dst.Path, src.Path, true, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to expand tar file: %w", err)
		}

		info, err := os.Stat(destination)
		if err != nil {
			return nil, fmt.Errorf("failed to get file info: %w", err)
		}

		return &file.FileMetadata{
			Size:      info.Size(),
			Path:      destination,
			Timestamp: info.ModTime(),
		}, nil
	}

	// If it's a directory, call copyDirectory, otherwise call copyFile
	if sourceKind.IsDir() {
		return f.copyDirectory(ctx, src.Path, destination)
	} else {
		return f.copyFile(ctx, src.Path, destination)
	}
}

func (f *FileGatherer) copyFile(ctx context.Context, source, destination string) (metadata.Metadata, error) {
	src, err := url.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source URI: %w", err)
	}
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("error copying file: %w", ctx.Err())
	default:
	}

	// Open the source file.
	srcFile, err := os.Open(filepath.Clean(src.Path))
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Parse the destination URI.
	destFile, err := url.Parse(destination)
	if err != nil {
		return nil, fmt.Errorf("failed to parse destination URI: %w", err)
	}

	// Create the appropriate Saver to handle storing the data.
	saver, err := saver.NewSaver("file")
	if err != nil {
		return nil, fmt.Errorf("failed to create saver: %w", err)
	}

	// Save the file to the destination.
	if err := saver.Save(ctx, srcFile, destination); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Get the file info
	info, err := os.Stat(destFile.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Calculate the SHA256 hash of the file
	fileSha, err := getFileSha(destFile.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate file SHA: %w", err)
	}

	return &file.FileMetadata{
		Size:      info.Size(),
		Path:      destination,
		Timestamp: info.ModTime(),
		SHA:       fileSha,
	}, nil
}

// copyDirectory copies a directory from the source path to the destination path.
// It walks through the directory tree, creates the corresponding directories in the destination path,
// and copies each file in the directory to the destination path.
// It limits the number of concurrent operations to 10 to avoid overwhelming system resources.
// It returns the metadata of the copied directory and any error encountered.
func (f *FileGatherer) copyDirectory(ctx context.Context, source, destination string) (metadata.Metadata, error) {
	src, err := url.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source URI: %w", err)
	}
	dst, err := url.Parse(destination)
	if err != nil {
		return nil, fmt.Errorf("failed to parse destination URI: %w", err)
	}

	errChan := make(chan error, 100) // Increased buffer size
	done := make(chan bool)
	semaphore := make(chan struct{}, 10) // Limit to 10 concurrent operations

	var wg sync.WaitGroup // Using a WaitGroup to manage concurrency

	go func() {
		defer close(done)
		err = filepath.Walk(src.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("failed to walk path: %w", err)
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			relPath, err := filepath.Rel(src.Path, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}

			destPath := filepath.Join(dst.Path, relPath)
			if info.IsDir() {
				if err := os.MkdirAll(destPath, 0755); err != nil {
					return fmt.Errorf("failed to create directory: %w", err)
				}
			} else {
				semaphore <- struct{}{}
				wg.Add(1)
				go func() {
					defer func() {
						<-semaphore
						wg.Done()
					}()
					srcFile, err := os.Open(filepath.Clean(path))
					if err != nil {
						errChan <- err
						return
					}
					defer srcFile.Close()

					saver, err := saver.NewSaver("file")
					if err != nil {
						errChan <- err
						return
					}

					if err := saver.Save(ctx, srcFile, destPath); err != nil {
						errChan <- err
						return
					}
				}()
			}
			return nil
		})
		if err != nil {
			errChan <- err
		}

		wg.Wait()      // Wait for all goroutines to finish
		close(errChan) // Close the channel safely after all sends are done
	}()

	// Handle errors and completion
	for err := range errChan {
		if err != nil {
			return nil, fmt.Errorf("failed to copy directory: %w", err)
		}
	}
	<-done
	return &file.DirectoryMetadata{
		Path:      dst.Path,
		Timestamp: time.Now(),
	}, nil
}

// getFileSha calculates the SHA256 hash of a file located at the given path.
// It returns the hexadecimal representation of the hash and any error encountered.
// If the file cannot be opened or an error occurs while calculating the hash, an empty string and the error are returned.
// The file is closed before returning.
func getFileSha(path string) (string, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to calculate file SHA: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
