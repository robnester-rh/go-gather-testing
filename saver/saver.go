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

// Package saver provides functionality for saving data to a destination.
// It defines an interface Saver for saving data and a function NewSaver for creating a Saver instance based on the destination protocol.
//
// The Saver interface defines a single method Save, which takes a context.Context, an io.Reader containing the data to be saved,
// and a destination string specifying the destination where the data should be saved. It returns an error if the save operation fails.
//
// The NewSaver function takes a protocol string as input and returns a Saver instance based on the specified protocol.
// Currently, the only supported protocol is "file", which creates a FileSaver instance for saving data to a file.
// If an unsupported protocol is provided, NewSaver returns an error.
//
// Example usage:
//
//	s, err := saver.NewSaver("file")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	err = s.Save(context.Background(), data, "/path/to/file.txt")
//	if err != nil {
//	    log.Fatal(err)
//	}
package saver

import (
	"context"
	"fmt"
	"io"

	"github.com/enterprise-contract/go-gather/saver/file"
)

// Saver is an interface for saving data to a destination.
type Saver interface {
	Save(ctx context.Context, data io.Reader, destination string) error
}

// NewSaver returns a Saver instance based on the destination protocol.
func NewSaver(protocol string) (Saver, error) {
	switch protocol {
	case "file", "FileURI":
		return &file.FileSaver{}, nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}
