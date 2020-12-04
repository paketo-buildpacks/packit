package internal

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/docker/distribution/reference"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type Image struct {
	Name    string
	Path    string
	Version string
}

func FindLatestImage(uri string) (Image, error) {
	named, err := reference.ParseNormalizedNamed(uri)
	if err != nil {
		return Image{}, fmt.Errorf("failed to parse image reference %q: %w", uri, err)
	}

	repo, err := name.NewRepository(reference.Path(named))
	if err != nil {
		return Image{}, fmt.Errorf("failed to parse image repository: %w", err)
	}

	repo.Registry, err = name.NewRegistry(reference.Domain(named))
	if err != nil {
		return Image{}, fmt.Errorf("failed to parse image registry: %w", err)
	}

	tags, err := remote.List(repo, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return Image{}, fmt.Errorf("failed to list tags: %w", err)
	}

	var versions []*semver.Version
	for _, tag := range tags {
		version, err := semver.NewVersion(tag)
		if err != nil {
			continue
		}

		versions = append(versions, version)
	}

	sort.Sort(semver.Collection(versions))

	return Image{
		Name:    named.Name(),
		Path:    reference.Path(named),
		Version: versions[len(versions)-1].String(),
	}, nil
}
