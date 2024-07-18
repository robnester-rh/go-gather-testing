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
