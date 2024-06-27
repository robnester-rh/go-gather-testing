package oci

type OCIMetadata struct {
	Digest string
}

func (o OCIMetadata) Get() map[string]any {
	return map[string]any{
		"digest": o.Digest,
	}
}
