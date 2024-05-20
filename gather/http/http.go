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

// Package http provides methods for gathering HTTP resources.
// This package implements the Gatherer interface and provides methods for downloading files from HTTP sources,
// retrieving metadata of the downloaded file, and handling HTTP requests.
//
// The HTTPGatherer struct represents an HTTP gatherer and contains a http.Client for making HTTP requests.
// It implements the Gatherer interface's Gather method to download files from HTTP sources and return the metadata of the downloaded file.
//
// Example usage:
//
//	httpGatherer := http.NewHTTPGatherer()
//	metadata, err := httpGatherer.Gather(context.Background(), "http://example.com/file.txt", "/path/to/destination/with/optional/filename.txt")
//	if err != nil {
//	  fmt.Println("Error:", err)
//	  return
//	}
//	fmt.Println("Downloaded file metadata:", metadata)
//
// Note: The Gather method uses the http.Client's default timeout of 15 seconds for the HTTP requests.
// You can customize the timeout by modifying the http.Client's Timeout field before calling the Gather method.
package http

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	gogather "github.com/enterprise-contract/go-gather"
	"github.com/enterprise-contract/go-gather/metadata"
	httpMetadata "github.com/enterprise-contract/go-gather/metadata/http"
	"github.com/enterprise-contract/go-gather/saver"
)

type HTTPGatherer struct {
	Client http.Client
}

func NewHTTPGatherer() *HTTPGatherer {
	return &HTTPGatherer{
		Client: http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (h *HTTPGatherer) Gather(ctx context.Context, source, destination string) (metadata.Metadata, error) {

	// Parse source
	u, err := url.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("error parsing source URI: %w", err)
	}

	// Check if the source scheme is provided
	if u.Scheme == "" {
		return nil, fmt.Errorf("no source scheme provided")
	}

	// Check if the source path is provided
	if u.Path == "" {
		return nil, fmt.Errorf("specify a path to a file to download")
	}

	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", source, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	req.Header.Set("User-Agent", "Go-Gather")
	
	// Send the HTTP request
	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error downloading file: %w", err)
	}
	defer resp.Body.Close()

	// Check if the response was successful
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response code error: %d", resp.StatusCode)
	}

	// Determine the destination type
	scheme, err := gogather.ClassifyURI(destination)
	if err != nil {
		return nil, fmt.Errorf("error determining destination type: %w", err)
	}

	// Check if the destination has a trailing slash.
	// If it does, append the source filename to the destination path.
	if destination[len(destination)-1] == '/' {
		destination = filepath.Join(destination, filepath.Base(u.Path))
		fmt.Println(destination)
	}

	// Create a new saver based on the destination scheme
	s, err := saver.NewSaver(scheme.String())
	if err != nil {
		return nil, fmt.Errorf("error creating saver: %w", err)
	}

	// Save the downloaded file
	err = s.Save(ctx, resp.Body, destination)
	if err != nil {
		if strings.Contains(err.Error(), "is a directory") {
			destination = filepath.Join(destination, filepath.Base(u.Path))
			err = s.Save(ctx, resp.Body, destination)
			if err != nil {
				return nil, fmt.Errorf("error saving file: %w", err)
			}
		}
	}

	// Return the metadata of the downloaded file
	m := httpMetadata.HTTPMetadata{
		StatusCode:    resp.StatusCode,
		ContentLength: resp.ContentLength,
		Destination:   destination,
		Headers:       resp.Header,
	}
	return m, nil
}
