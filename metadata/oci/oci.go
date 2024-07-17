package oci

type OCIMetadata struct {
	Digest string
}

func (o OCIMetadata) Get() map[string]any {
	return map[string]any{
		"digest": o.Digest,
	}
}

// GetDigest returns the digest of the artifact.
func (o OCIMetadata) GetDigest() string {
	return o.Digest
}
