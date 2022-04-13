package packit

import "github.com/paketo-buildpacks/packit/v2/planning"

// BuildPlan is a representation of the Build Plan as specified in the
// specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#build-plan-toml.
// The BuildPlan allows buildpacks to indicate what dependencies they provide
// or require.
type BuildPlan = BuildPlanOf[interface{}]

type BuildPlanOf[T interface{} | planning.Metadata] struct {
	// Provides is a list of BuildPlanProvisions that are provided by this
	// buildpack.
	Provides []BuildPlanProvision `toml:"provides"`

	// Requires is a list of BuildPlanRequirements that are required by this
	// buildpack.
	Requires []BuildPlanRequirementOf[T] `toml:"requires"`

	// Or is a list of additional BuildPlans that may be selected by the
	// lifecycle
	Or []BuildPlanOf[T] `toml:"or,omitempty"`
}

// BuildPlanProvision is a representation of a dependency that can be provided
// by a buildpack.
type BuildPlanProvision struct {
	// Name is the identifier whereby buildpacks can coordinate that a dependency
	// is provided or required.
	Name string `toml:"name"`
}

type BuildPlanRequirementOf[T interface{} | planning.Metadata] struct {
	// Name is the identifier whereby buildpacks can coordinate that a dependency
	// is provided or required.
	Name string `toml:"name"`

	// Metadata is an unspecified field allowing buildpacks to communicate extra
	// details about their requirement. Examples of this type of metadata might
	// include details about what source was used to decide the version
	// constraint for a requirement.
	Metadata T `toml:"metadata"`
}

type BuildPlanRequirement = BuildPlanRequirementOf[interface{}]
