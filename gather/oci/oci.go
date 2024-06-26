package oci

import (
	"context"
	"fmt"
	"os"
	"strings"

	r "github.com/enterprise-contract/go-gather/gather/oci/internal/registry"

	"github.com/enterprise-contract/go-gather/metadata"
	"github.com/enterprise-contract/go-gather/metadata/oci"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

// OCIGatherer is a struct that implements the Gatherer interface
// and provides methods for gathering from OCI.
type OCIGatherer struct{}

// Gather copies a file or directory from the source path to the destination path.
// It returns the metadata of the gathered file or directory and any error encountered.
// Portions of this file are derivative from the open-policy-agent/conftest project.
func (f *OCIGatherer) Gather(ctx context.Context, source, destination string) (metadata.Metadata, error) {
	// Parse the source URI
	repo := ociURLParse(source)

	// Get the artifact reference
	ref, err := registry.ParseReference(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: %w", err)
	}

	// If the reference is empty, set it to "latest"
	if ref.Reference == "" {
		ref.Reference = "latest"
		repo = ref.String()
	}

	// Create the repository client
	src, err := remote.NewRepository(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository client: %w", err)
	}

	// Setup the client for the repository
	if err := r.SetupClient(src); err != nil {
		return nil, fmt.Errorf("failed to setup repository client: %w", err)
	}

	// Create the destination directory
	if err := os.MkdirAll(destination, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file store
	fileStore, err := file.New(destination)
	if err != nil {
		return nil, fmt.Errorf("file store: %w", err)
	}
	defer fileStore.Close()

	// Copy the artifact to the file store
	a, err := oras.Copy(ctx, src, repo, fileStore, "", oras.DefaultCopyOptions)
	if err != nil {
		return nil, fmt.Errorf("pulling policy: %w", err)
	}

	m := &oci.OCIMetadata{
		Digest: a.Digest.String(),
	}
	return m, nil
}

func ociURLParse(source string) string {
	scheme, src, found := strings.Cut(source, "://")
	if !found {
		src = scheme
	}
	return src
}