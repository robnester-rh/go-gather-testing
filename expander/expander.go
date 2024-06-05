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
	"fmt"
	"io"
	"os"
	"strings"
)

// Expander is an interface which defines the methods that an expander must implement in order expand a type
type Expander interface {
	Expand(src, dst string, dir bool, mode os.FileMode) error
}

// BaseExpanders creates the set of base expanders that are used to expand the different types of files
func BaseExpanders(filesLimit int, fileSizeLimit int64) map[string]Expander {
	return map[string]Expander{
		"tar": &TarExpander{},
	}
}

// containsDotDot checks if the filepath value v contains a ".." entry.
// This will check filepath components by splitting along / or \. This
// function is copied directly from the Go net/http implementation.
func containsDotDot(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}
	for _, ent := range strings.FieldsFunc(v, isSlash) {
		if ent == ".." {
			return true
		}
	}
	return false
}

func isSlash(r rune) bool { return r == '/' || r == '\\' }

// copyReader copies a reader to a file. If fileSizeLimit is greater than 0, it will limit the size of the file.
func copyReader(src io.Reader, dst string, mode os.FileMode, fileSizeLimit int64) error {
	dstF, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", dst, err)
	}
	defer dstF.Close()

	if fileSizeLimit > 0 {
		src = io.LimitReader(src, fileSizeLimit)
	}

	_, err = io.Copy(dstF, src)
	if err != nil {
		return fmt.Errorf("failed to copy file %s: %w", dst, err)
	}

	return os.Chmod(dst, mode)
}
