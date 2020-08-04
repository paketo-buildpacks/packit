package packit

import (
	"fmt"
	"os"
)

// LayerType defines the set of layer types that can be declared according to
// the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#layer-types.
type LayerType uint8

const (
	// BuildLayer indicates that the layer will be made available during the
	// build phase.
	BuildLayer LayerType = iota

	// LaunchLayer indicates that the layer will be made available during the
	// launch phase.
	LaunchLayer

	// CacheLayer indicates that the layer will be cached and made available to
	// the buildpack on subsequent rebuilds.
	CacheLayer
)

// Layer provides a representation of a layer managed by a buildpack as
// described by the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#layers.
type Layer struct {
	// Path is the absolute location of the layer on disk.
	Path string `toml:"-"`

	// Name is the descriptive name of the layer.
	Name string `toml:"-"`

	// Build indicates whether the layer is available to subsequent buildpacks
	// during their build phase according to the specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#build-layers.
	Build bool `toml:"build"`

	// Launch indicates whether the layer is exported into the application image
	// and made available during the launch phase according to the specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#launch-layers.
	Launch bool `toml:"launch"`

	// Cache indicates whether the layer is persisted and made available to
	// subsequent builds of the same application according to the specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#launch-layers
	// and
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#build-layers.
	Cache bool `toml:"cache"`

	// SharedEnv is the set of environment variables attached to the layer and
	// made available during both the build and launch phases according to the
	// specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#provided-by-the-buildpacks.
	SharedEnv Environment `toml:"-"`

	// BuildEnv is the set of environment variables attached to the layer and
	// made available during the build phase according to the specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#provided-by-the-buildpacks.
	BuildEnv Environment `toml:"-"`

	// LaunchEnv is the set of environment variables attached to the layer and
	// made available during the launch phase according to the specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#provided-by-the-buildpacks.
	LaunchEnv Environment `toml:"-"`

	// Metadata is an unspecified field allowing buildpacks to communicate extra
	// details about the layer. Examples of this type of metadata might include
	// details about what versions of software are included in the layer such
	// that subsequent builds can inspect that metadata and choose to reuse the
	// layer if suitable. The Metadata field ultimately fills the metadata field
	// of the Layer Content Metadata TOML file according to the specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#layer-content-metadata-toml.
	Metadata map[string]interface{} `toml:"metadata"`
}

// Reset clears the state of a layer such that the layer can be replaced with
// new content and metadata. It clears all environment variables, and removes
// the content of the layer directory on disk.
func (l *Layer) Reset() error {
	l.SharedEnv = Environment{}
	l.BuildEnv = Environment{}
	l.LaunchEnv = Environment{}
	l.Metadata = nil

	err := os.RemoveAll(l.Path)
	if err != nil {
		return fmt.Errorf("error could not remove file: %s", err)
	}

	err = os.MkdirAll(l.Path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error could not create directory: %s", err)
	}

	return nil
}
