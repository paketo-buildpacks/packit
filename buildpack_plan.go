package packit

// BuildpackPlan is a representation of the buildpack plan provided by the
// lifecycle and defined in the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#buildpack-plan-toml.
// It is also used to return a set of refinements to the plan at the end of the
// build phase.
type BuildpackPlan struct {
	// Entries is a list of BuildpackPlanEntry fields that are declared in the
	// buildpack plan TOML file.
	Entries []BuildpackPlanEntry `toml:"entries"`
}

// BuildpackPlanEntry is a representation of a single buildpack plan entry
// specified by the lifecycle.
type BuildpackPlanEntry struct {
	// Name is the name of the dependency the the buildpack should provide.
	Name string `toml:"name"`

	// Metadata is an unspecified field allowing buildpacks to communicate extra
	// details about their requirement. Examples of this type of metadata might
	// include details about what source was used to decide the version
	// constraint for a requirement.
	Metadata map[string]interface{} `toml:"metadata"`
}
