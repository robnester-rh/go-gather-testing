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

package expander

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// untar is a helper function that untars a tarball to a destination directory
func untar(input io.Reader, dst, src string, dir bool, umask os.FileMode, fileSizeLimit int64, filesLimit int) error {
	tarReader := tar.NewReader(input)
	finished := false

	dirHeaders := []*tar.Header{}
	now := time.Now()

	var (
		fileSize   int64
		filesCount int
	)

	for {
		if filesLimit > 0 {
			filesCount++
			if filesCount > filesLimit {
				return fmt.Errorf("tar file contains more files than the %d allowed: %d", filesCount, filesLimit)
			}
		}

		header, err := tarReader.Next()
		if err == io.EOF {
			if !finished {
				// Empty archive
				return fmt.Errorf("tar file is empty: %s", src)
			}
			break
		}

		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeXGlobalHeader || header.Typeflag == tar.TypeXHeader {
			continue
		}

		fPath := dst

		if dir {
			if containsDotDot(header.Name) {
				return fmt.Errorf("tar file (%s) would escape destination directory", header.Name)
			}

			fPath = filepath.Join(dst, header.Name) // nolint:gosec
		}

		fileInfo := header.FileInfo()
		fileSize += fileInfo.Size()

		if fileSizeLimit > 0 && fileSize > fileSizeLimit {
			return fmt.Errorf("tar file size exceeds the %d limit: %d", fileSizeLimit, fileSize)
		}

		if fileInfo.IsDir() {
			if !dir {
				return fmt.Errorf("expected a file (%s), got a directory: %s", src, fPath)
			}

			if err := os.MkdirAll(fPath, umask); err != nil {
				return fmt.Errorf("failed to create directory (%s): %s", fPath, err)
			}

			dirHeaders = append(dirHeaders, header)

			continue
		} else {
			destPath := filepath.Dir(fPath)

			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				if err := os.MkdirAll(destPath, umask); err != nil {
					return fmt.Errorf("failed to create directory (%s): %s", destPath, err)
				}
			}
		}

		if !dir && finished {
			return fmt.Errorf("tar file contains more than one file: %s", src)
		}

		finished = true

		err = copyReader(tarReader, fPath, umask, fileSizeLimit)
		if err != nil {
			return err
		}

		aTime, mTime := now, now

		if header.AccessTime.Unix() > 0 {
			aTime = header.AccessTime
		}

		if header.ModTime.Unix() > 0 {
			mTime = header.ModTime
		}

		if err := os.Chtimes(fPath, aTime, mTime); err != nil {
			return fmt.Errorf("failed to change file times (%s): %s", fPath, err)
		}
	}

	for _, dirHeader := range dirHeaders {
		if containsDotDot(dirHeader.Name) {
			return fmt.Errorf("tar file (%s) would escape destination directory", dirHeader.Name)
		}
		path := filepath.Join(dst, dirHeader.Name) // nolint:gosec
		// Chmod the directory
		if err := os.Chmod(path, dirHeader.FileInfo().Mode()); err != nil {
			return fmt.Errorf("failed to change directory permissions (%s): %s", path, err)
		}

		// Set the access and modification times
		aTime, mTime := now, now

		if dirHeader.AccessTime.Unix() > 0 {
			aTime = dirHeader.AccessTime
		}
		if dirHeader.ModTime.Unix() > 0 {
			mTime = dirHeader.ModTime
		}
		if err := os.Chtimes(path, aTime, mTime); err != nil {
			return fmt.Errorf("failed to change directory times (%s): %s", path, err)
		}
	}
	return nil
}

type TarExpander struct {
	FileSizeLimit int64
	FilesLimit    int
}

func (t *TarExpander) Expand(dst, src string, dir bool, umask os.FileMode) error {
	if !dir {
		err := os.MkdirAll(dst, umask)
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	return untar(f, dst, src, dir, umask, t.FileSizeLimit, t.FilesLimit)
}
