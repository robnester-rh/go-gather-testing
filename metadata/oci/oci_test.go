package oci

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOCIMetadata_Get(t *testing.T) {
	o := OCIMetadata{Digest: "fa93b01658e3a5a1686dc3ae55f170d8de487006fb53a28efcd12ab0710a2e5f"}
	expected := map[string]any{
		"digest": "fa93b01658e3a5a1686dc3ae55f170d8de487006fb53a28efcd12ab0710a2e5f",
	}
	result := o.Get()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected Get() to return %v, but got %v", expected, result)
	}
}

func TestOCIMetadata_GetDigest(t *testing.T) {
	o := OCIMetadata{Digest: "fa93b01658e3a5a1686dc3ae55f170d8de487006fb53a28efcd12ab0710a2e5f"}
	expected := "fa93b01658e3a5a1686dc3ae55f170d8de487006fb53a28efcd12ab0710a2e5f"
	result := o.GetDigest()
	assert.Equal(t, expected, result, "Expected GetDigest() to return %s, but got %s", expected, result)
}
