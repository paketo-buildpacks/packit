package packit

// BuildpackInfo is a representation of the basic information for a buildpack
// provided in its buildpack.toml file as described in the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#buildpacktoml-toml.
type BuildpackInfo struct {
	// ID is the identifier specified in the `buildpack.id` field of the buildpack.toml.
	ID string `toml:"id"`

	// Name is the identifier specified in the `buildpack.name` field of the buildpack.toml.
	Name string `toml:"name"`

	// Version is the identifier specified in the `buildpack.version` field of the buildpack.toml.
	Version string `toml:"version"`
}
