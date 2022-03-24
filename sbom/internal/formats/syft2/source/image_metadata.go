package source

import "github.com/anchore/syft/syft/source"

func ConvertImageMetadata(metadata source.ImageMetadata) ImageMetadata {
	// TODO: (packit) Is there a cleaner way to unpack the new struct into the legacy one?
	var layers []LayerMetadata
	for _, l := range metadata.Layers {
		layers = append(layers, LayerMetadata{
			MediaType: l.MediaType,
			Digest:    l.Digest,
			Size:      l.Size,
		})
	}

	// Create RepoDigests slice this way to ensure that it's encoded
	// as an empty array (not null) if empty
	repoDigests := make([]string, 0)
	if metadata.RepoDigests != nil {
		repoDigests = metadata.RepoDigests
	}

	return ImageMetadata{
		UserInput:      metadata.UserInput,
		ID:             metadata.ID,
		ManifestDigest: metadata.ManifestDigest,
		MediaType:      metadata.MediaType,
		Tags:           metadata.Tags,
		Size:           metadata.Size,
		Layers:         layers,
		RawManifest:    metadata.RawManifest,
		RawConfig:      metadata.RawConfig,
		RepoDigests:    repoDigests,
	}
}

// Source: https://github.com/anchore/syft/blob/dfefd2ea4e9d44187b4f861bc202970247dd34c8/syft/source/image_metadata.go
// ImageMetadata represents all static metadata that defines what a container image is. This is useful to later describe
// "what" was cataloged without needing the more complicated stereoscope Image objects or FileResolver objects.
type ImageMetadata struct {
	UserInput      string          `json:"userInput"`
	ID             string          `json:"imageID"`
	ManifestDigest string          `json:"manifestDigest"`
	MediaType      string          `json:"mediaType"`
	Tags           []string        `json:"tags"`
	Size           int64           `json:"imageSize"`
	Layers         []LayerMetadata `json:"layers"`
	RawManifest    []byte          `json:"manifest"`
	RawConfig      []byte          `json:"config"`
	RepoDigests    []string        `json:"repoDigests"`
}

// LayerMetadata represents all static metadata that defines what a container image layer is.
type LayerMetadata struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}
