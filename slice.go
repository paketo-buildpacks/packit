package packit

// Slice represents a layer of the working directory to be exported during the
// export phase. These slices help to optimize data transfer for files that are
// commonly shared across applications.  Slices are described in the layers
// section of the buildpack spec:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#layers.  The slice
// fields are described in the specification of the launch.toml file:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#launchtoml-toml.
type Slice struct {
	Paths []string `toml:"paths"`
}
