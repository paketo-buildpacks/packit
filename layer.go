package packit

import (
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/pelletier/go-toml"
)

// Layer provides a representation of a layer managed by a buildpack as
// described by the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#layers.
type Layer struct {
	// Path is the absolute location of the layer on disk.
	Path string

	// Name is the descriptive name of the layer.
	Name string

	// Build indicates whether the layer is available to subsequent buildpacks
	// during their build phase according to the specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#build-layers.
	Build bool

	// Launch indicates whether the layer is exported into the application image
	// and made available during the launch phase according to the specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#launch-layers.
	Launch bool

	// Cache indicates whether the layer is persisted and made available to
	// subsequent builds of the same application according to the specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#launch-layers
	// and
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#build-layers.
	Cache bool

	// SharedEnv is the set of environment variables attached to the layer and
	// made available during both the build and launch phases according to the
	// specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#provided-by-the-buildpacks.
	SharedEnv Environment

	// BuildEnv is the set of environment variables attached to the layer and
	// made available during the build phase according to the specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#provided-by-the-buildpacks.
	BuildEnv Environment

	// LaunchEnv is the set of environment variables attached to the layer and
	// made available during the launch phase according to the specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#provided-by-the-buildpacks.
	LaunchEnv Environment

	// ProcessLaunchEnv is a map of environment variables attached to the layer and
	// made available to specified proccesses in the launch phase accoring to the specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#provided-by-the-buildpacks
	ProcessLaunchEnv map[string]Environment

	// Metadata is an unspecified field allowing buildpacks to communicate extra
	// details about the layer. Examples of this type of metadata might include
	// details about what versions of software are included in the layer such
	// that subsequent builds can inspect that metadata and choose to reuse the
	// layer if suitable. The Metadata field ultimately fills the metadata field
	// of the Layer Content Metadata TOML file according to the specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#layer-content-metadata-toml.
	Metadata map[string]interface{}

	// SBOM is a type that implements SBOMFormatter and declares the formats that
	// bill-of-materials should be output for the layer SBoM.
	SBOM SBOMFormatter
}

// Reset clears the state of a layer such that the layer can be replaced with
// new content and metadata. It clears all environment variables, and removes
// the content of the layer directory on disk.
func (l Layer) Reset() (Layer, error) {
	l.Build = false
	l.Launch = false
	l.Cache = false

	l.SharedEnv = Environment{}
	l.BuildEnv = Environment{}
	l.LaunchEnv = Environment{}
	l.ProcessLaunchEnv = make(map[string]Environment)
	l.Metadata = nil

	err := os.RemoveAll(l.Path)
	if err != nil {
		return Layer{}, fmt.Errorf("error could not remove file: %s", err)
	}

	err = os.MkdirAll(l.Path, os.ModePerm)
	if err != nil {
		return Layer{}, fmt.Errorf("error could not create directory: %s", err)
	}

	return l, nil
}

type formattedLayer struct {
	layer Layer
	api   *semver.Version
}

func (l formattedLayer) MarshalTOML() ([]byte, error) {
	layer := map[string]interface{}{
		"metadata": l.layer.Metadata,
	}

	apiV06, _ := semver.NewVersion("0.6")
	if l.api.LessThan(apiV06) {
		layer["build"] = l.layer.Build
		layer["launch"] = l.layer.Launch
		layer["cache"] = l.layer.Cache
	} else {
		layer["types"] = map[string]bool{
			"build":  l.layer.Build,
			"launch": l.layer.Launch,
			"cache":  l.layer.Cache,
		}
	}

	return toml.Marshal(layer)
}
