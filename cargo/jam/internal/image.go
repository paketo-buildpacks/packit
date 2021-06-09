package internal

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
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
		version, err := semver.StrictNewVersion(tag)
		if err != nil {
			continue
		}
		if version.Prerelease() != "" {
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

func FindLatestBuildImage(runURI, buildURI string) (Image, error) {
	runNamed, err := reference.ParseNormalizedNamed(runURI)
	if err != nil {
		return Image{}, fmt.Errorf("failed to parse run image reference %q: %w", runURI, err)
	}

	tagged, ok := runNamed.(reference.Tagged)
	if !ok {
		return Image{}, fmt.Errorf("expected the run image to be tagged but it was not")
	}

	suffix := fmt.Sprintf("-%s", tagged.Tag())

	buildNamed, err := reference.ParseNormalizedNamed(buildURI)
	if err != nil {
		return Image{}, fmt.Errorf("failed to parse build image reference %q: %w", buildURI, err)
	}

	repo, err := name.NewRepository(reference.Path(buildNamed))
	if err != nil {
		return Image{}, fmt.Errorf("failed to parse build image repository: %w", err)
	}

	repo.Registry, err = name.NewRegistry(reference.Domain(buildNamed))
	if err != nil {
		return Image{}, fmt.Errorf("failed to parse build image registry: %w", err)
	}

	tags, err := remote.List(repo, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return Image{}, fmt.Errorf("failed to list tags: %w", err)
	}

	var versions []*semver.Version
	for _, tag := range tags {
		if !strings.HasSuffix(tag, suffix) {
			continue
		}

		version, err := semver.StrictNewVersion(strings.TrimSuffix(tag, suffix))
		if err != nil {
			continue
		}

		versions = append(versions, version)
	}

	sort.Sort(semver.Collection(versions))

	return Image{
		Name:    buildNamed.Name(),
		Path:    reference.Path(buildNamed),
		Version: fmt.Sprintf("%s%s", versions[len(versions)-1].String(), suffix),
	}, nil
}

func GetBuildpackageID(uri string) (string, error) {
	ref, err := name.ParseReference(uri)
	if err != nil {
		return "", err
	}

	image, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return "", err
	}

	cfg, err := image.ConfigFile()
	if err != nil {
		return "", err
	}

	type BuildpackageMetadata struct {
		BuildpackageID string `json:"id"`
	}
	var metadataString string
	var ok bool
	if metadataString, ok = cfg.Config.Labels["io.buildpacks.buildpackage.metadata"]; !ok {
		return "", fmt.Errorf("could not get buildpackage id: image %s has no label 'io.buildpacks.buildpackage.metadata'", uri)
	}

	metadata := BuildpackageMetadata{}

	err = json.Unmarshal([]byte(metadataString), &metadata)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal buildpackage metadata")
	}
	return metadata.BuildpackageID, nil
}
