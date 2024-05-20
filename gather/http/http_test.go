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

package http

import (
	"context"
	"fmt"
	h "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/enterprise-contract/go-gather/metadata/http"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPGatherer(t *testing.T) {
	gatherer := NewHTTPGatherer()

	// Verify the timeout value
	expectedTimeout := 15 * time.Second
	if gatherer.Client.Timeout != expectedTimeout {
		t.Errorf("unexpected timeout value: got %v, want %v", gatherer.Client.Timeout, expectedTimeout)
	}
}

func TestHTTPGatherer_Gather_WithTrailingSlash(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	tempDir = fmt.Sprintf("%s/", tempDir)

	// Create a mock HTTP server
	mockServer := httptest.NewServer(h.HandlerFunc(func(w h.ResponseWriter, r *h.Request) {
		// Set the Content-Type header
		w.Header().Set("Content-Type", "text/plain")

		// Write the foo.bar content
		fmt.Fprint(w, "Hello, World!")
	}))
	defer mockServer.Close()

	// Create a new HTTPGatherer instance
	gatherer := NewHTTPGatherer()

	// Call the Gather method with the mock server URL and the temporary directory
	m, err := gatherer.Gather(context.Background(), fmt.Sprintf("%s/foo.bar", mockServer.URL), fmt.Sprintf("%s/", tempDir))
	if err != nil {
		t.Fatal(err)
	}

	// Verify the metadata
	expectedStatusCode := h.StatusOK
	if m.(http.HTTPMetadata).StatusCode != expectedStatusCode {
		t.Errorf("unexpected status code: got %d, want %d", m.(http.HTTPMetadata).StatusCode, expectedStatusCode)
	}

	expectedContentLength := int64(13)
	if m.(http.HTTPMetadata).ContentLength != expectedContentLength {
		t.Errorf("unexpected content length: got %d, want %d", m.(http.HTTPMetadata).ContentLength, expectedContentLength)
	}

	expectedDestination := fmt.Sprintf("%sfoo.bar", tempDir)
	if m.(http.HTTPMetadata).Destination != expectedDestination {
		t.Errorf("unexpected destination: got %s, want %s", m.(http.HTTPMetadata).Destination, expectedDestination)
	}

	// Verify the downloaded file
	filePath := filepath.Join(tempDir, "/foo.bar")
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	expectedFileContent := "Hello, World!"
	if string(fileContent) != expectedFileContent {
		t.Errorf("unexpected file content: got %s, want %s", string(fileContent), expectedFileContent)
	}
}
// TestHTTPGatherer_Gather_WithoutTrailingSlash tests the Gather method with a destination that does not have a trailing slash.
func TestHTTPGatherer_Gather_WithoutTrailingSlash(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock HTTP server
	mockServer := httptest.NewServer(h.HandlerFunc(func(w h.ResponseWriter, r *h.Request) {
		// Set the Content-Type header
		w.Header().Set("Content-Type", "text/plain")

		// Write the foo.bar content
		fmt.Fprint(w, "Hello, World!")
	}))
	defer mockServer.Close()

	// Create a new HTTPGatherer instance
	gatherer := NewHTTPGatherer()

	// Call the Gather method with the mock server URL and the temporary directory
	m, err := gatherer.Gather(context.Background(), fmt.Sprintf("%s/foo.bar", mockServer.URL), tempDir)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the metadata
	expectedStatusCode := h.StatusOK
	if m.(http.HTTPMetadata).StatusCode != expectedStatusCode {
		t.Errorf("unexpected status code: got %d, want %d", m.(http.HTTPMetadata).StatusCode, expectedStatusCode)
	}

	expectedContentLength := int64(13)
	if m.(http.HTTPMetadata).ContentLength != expectedContentLength {
		t.Errorf("unexpected content length: got %d, want %d", m.(http.HTTPMetadata).ContentLength, expectedContentLength)
	}

	expectedDestination := fmt.Sprintf("%s/foo.bar", tempDir)
	if m.(http.HTTPMetadata).Destination != expectedDestination {
		t.Errorf("unexpected destination: got %s, want %s", m.(http.HTTPMetadata).Destination, expectedDestination)
	}

	// Verify the downloaded file
	filePath := filepath.Join(tempDir, "/foo.bar")
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	expectedFileContent := "Hello, World!"
	if string(fileContent) != expectedFileContent {
		t.Errorf("unexpected file content: got %s, want %s", string(fileContent), expectedFileContent)
	}
}

// TestHTTPGatherer_Gather_ParseError tests the Gather method with a url.Parse error.
func TestHTTPGatherer_Gather_ParseError(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new HTTPGatherer instance
	gatherer := NewHTTPGatherer()

	// Call the Gather method with an unparsable source URI
	_, err = gatherer.Gather(context.Background(), ":", tempDir)
	if err == nil {
		t.Error("expected an error, but got nil")
	}
}

// TestHTTPGatherer_Gather_InvalidSource tests the Gather method with an invalid source URI.
func TestHTTPGatherer_Gather_InvalidSource(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new HTTPGatherer instance
	gatherer := NewHTTPGatherer()

	// Call the Gather method with an invalid source URI
	_, err = gatherer.Gather(context.Background(), "invalid-url", tempDir)
	if err == nil {
		t.Error("expected an error, but got nil")
	}
}

func TestHTTPGatherer_Gather_NewRequestWithContextError(t *testing.T) {
	mockServer := httptest.NewServer(h.HandlerFunc(func(w h.ResponseWriter, r *h.Request) {
		w.WriteHeader(h.StatusOK)
	}))
	defer mockServer.Close()

	h := &HTTPGatherer{}

	// Pass a nil context to the Gather method to force an error
	_, err := h.Gather(nil, fmt.Sprintf("%s/foo.bar", mockServer.URL), "/tmp")
	if err == nil {
		t.Fatal("expected an error but got nil")
	}

	expectedErrMsg := "error creating request: net/http: nil Context"
	if err.Error() != expectedErrMsg {
		t.Fatalf("expected error message %q but got %q", expectedErrMsg, err.Error())
	}
}

// TestHTTPGatherer_Gather_BadStatusCode tests the Gather method with a bad status code.
func TestHTTPGatherer_Gather_BadStatusCode(t *testing.T) {
	mockServer := httptest.NewServer(h.HandlerFunc(func(w h.ResponseWriter, r *h.Request) {
		w.WriteHeader(404)
	}))
	defer mockServer.Close()

	h := &HTTPGatherer{}

	_, err := h.Gather(context.Background(), fmt.Sprintf("%s/foo.bar", mockServer.URL), "/tmp")
	if err == nil {
		t.Fatal("expected an error but got nil")
	}

	expectedErrMsg := "response code error: 404"
	if err.Error() != expectedErrMsg {
		t.Fatalf("expected error message %q but got %q", expectedErrMsg, err.Error())
	}
}

// TestHTTPGatherer_Gather_HTTPError tests the Gather method with an HTTP error.
func TestHTTPGatherer_Gather_HTTPError(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock HTTP server that returns an error
	mockServer := httptest.NewServer(h.HandlerFunc(func(w h.ResponseWriter, r *h.Request) {
		w.WriteHeader(404)
	}))
	defer mockServer.Close()

	// Create a new HTTPGatherer instance
	gatherer := NewHTTPGatherer()

	// Call the Gather method with the mock server URL and the temporary directory
	_, err = gatherer.Gather(context.Background(), mockServer.URL, tempDir)
	if err == nil {
		t.Error("expected an error, but got nil")
	}
}

// TestHTTPGatherer_Gather_Client_Do_Error tests the Gather method with an error from http.Client.Do.
func TestHTTPGatherer_Gather_Client_Do_Error(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new HTTPGatherer instance with a custom timeout
	gatherer := &HTTPGatherer{
		Client: h.Client{
			Timeout: 1 * time.Nanosecond,
		},
	}

	// Create a mock HTTP server
	mockServer := httptest.NewServer(h.HandlerFunc(func(w h.ResponseWriter, r *h.Request) {
		// Return a sample response
		w.WriteHeader(h.StatusOK)
		fmt.Fprint(w, "Hello, World!")
	}))
	defer mockServer.Close()

	// Call the Gather method with the mock server URL and the temporary directory
	_, err = gatherer.Gather(context.Background(), fmt.Sprintf("%s/foo", mockServer.URL), tempDir)
	if err == nil {
		t.Error("expected an error, but got nil")
	}
	assert.ErrorContains(t, err, "context deadline exceeded (Client.Timeout exceeded while awaiting headers)")
}

// TestHTTPGatherer_Gather_ClassifyURI_Error tests the Gather method with an error from ClassifyURI.
func TestHTTPGatherer_Gather_ClassifyURI_Error(t *testing.T) {

	mockServer := httptest.NewServer(h.HandlerFunc(func(w h.ResponseWriter, r *h.Request) {
		w.WriteHeader(h.StatusOK)
	}))
	defer mockServer.Close()

	// Create a new HTTPGatherer instance
	gatherer := NewHTTPGatherer()

	// Call the Gather method with an invalid destination URI
	_, err := gatherer.Gather(context.Background(), fmt.Sprintf("%s/foo.bar", mockServer.URL), "foo://invalid-directory")
	if err == nil {
		t.Error("expected an error, but got nil")
	}
	assert.EqualError(t, err, "error determining destination type: unsupported source protocol: foo")
}
