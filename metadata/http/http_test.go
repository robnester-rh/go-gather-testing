package http

import (
	"reflect"
	"testing"
)

func TestHTTPMetadata_Get(t *testing.T) {
	// Create a sample HTTPMetadata instance
	metadata := HTTPMetadata{
		StatusCode:    200,
		ContentLength: 1024,
		Destination:   "https://example.com",
		Headers:       map[string][]string{"Content-Type": {"text/plain"}},
	}

	// Call the Get method
	result := metadata.Get()

	// Verify the expected values
	expected := map[string]interface{}{
		"statusCode":    200,
		"contentLength": int64(1024),
		"destination":   "https://example.com",
		"headers":       map[string][]string{"Content-Type": {"text/plain"}},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("unexpected result: got %v, want %v", result, expected)
	}
}
