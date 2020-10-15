package packit

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/packit/internal"
)

// DetectContext provides the contextual details that are made available by the
// buildpack lifecycle during the detect phase. This context is populated by
// the Detect function and passed to the DetectFunc during execution.
type DetectContext struct {
	// WorkingDir is the location of the application source code as provided by
	// the lifecycle.
	WorkingDir string

	// CNBPath is the absolute path location of the buildpack contents.
	// This path is useful for finding the buildpack.toml or any other
	// files included in the buildpack.
	CNBPath string

	// BuildpackInfo includes the details of the buildpack parsed from the
	// buildpack.toml included in the buildpack contents.
	BuildpackInfo BuildpackInfo

	// Stack is the value of the chosen stack. This value is populated from the
	// $CNB_STACK_ID environment variable.
	Stack string
}

// DetectFunc is the definition of a callback that can be invoked when the
// Detect function is executed. Buildpack authors should implement a DetectFunc
// that performs the specific detect phase operations for a buildpack.
type DetectFunc func(DetectContext) (DetectResult, error)

// DetectResult allows buildpack authors to indicate the result of the detect
// phase for a given buildpack. This result, returned in a DetectFunc callback,
// will be parsed and persisted by the Detect function and returned to the
// lifecycle at the end of the detect phase execution.
type DetectResult struct {
	// Plan is the set of Build Plan provisions and requirements that are
	// detected during the detect phase of the lifecycle.
	Plan BuildPlan
}

// BuildPlan is a representation of the Build Plan as specified in the
// specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#build-plan-toml.
// The BuildPlan allows buildpacks to indicate what dependencies they provide
// or require.
type BuildPlan struct {
	// Provides is a list of BuildPlanProvisions that are provided by this
	// buildpack.
	Provides []BuildPlanProvision `toml:"provides"`

	// Requires is a list of BuildPlanRequirements that are required by this
	// buildpack.
	Requires []BuildPlanRequirement `toml:"requires"`

	// Or is a list of additional BuildPlans that may be selected by the
	// lifecycle
	Or []BuildPlan `toml:"or,omitempty"`
}

// BuildPlanProvision is a representation of a dependency that can be provided
// by a buildpack.
type BuildPlanProvision struct {
	// Name is the identifier whereby buildpacks can coordinate that a dependency
	// is provided or required.
	Name string `toml:"name"`
}

type BuildPlanRequirement struct {
	// Name is the identifier whereby buildpacks can coordinate that a dependency
	// is provided or required.
	Name string `toml:"name"`

	// Version allows a requirement to include a constraint describing what
	// versions of the dependency are considered acceptable.
	//
	// Deprecated: Store version information in Metadata instead.
	Version string `toml:"version"`

	// Metadata is an unspecified field allowing buildpacks to communicate extra
	// details about their requirement. Examples of this type of metadata might
	// include details about what source was used to decide the version
	// constraint for a requirement.
	Metadata interface{} `toml:"metadata"`
}

// Detect is an implementation of the detect phase according to the Cloud
// Native Buildpacks specification. Calling this function with a DetectFunc
// will perform the detect phase process.
func Detect(f DetectFunc, options ...Option) {
	config := OptionConfig{
		exitHandler: internal.NewExitHandler(),
		args:        os.Args,
	}

	for _, option := range options {
		config = option(config)
	}

	dir, err := os.Getwd()
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

	cnbPath, ok := os.LookupEnv("CNB_BUILDPACK_DIR")
	if !ok {
		cnbPath = filepath.Clean(strings.TrimSuffix(config.args[0], filepath.Join("bin", "detect")))
	}

	var buildpackInfo struct {
		Buildpack BuildpackInfo `toml:"buildpack"`
	}
	_, err = toml.DecodeFile(filepath.Join(cnbPath, "buildpack.toml"), &buildpackInfo)
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

	result, err := f(DetectContext{
		WorkingDir:    dir,
		CNBPath:       cnbPath,
		BuildpackInfo: buildpackInfo.Buildpack,
		Stack:         os.Getenv("CNB_STACK_ID"),
	})
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

	file, err := os.OpenFile(config.args[2], os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		config.exitHandler.Error(err)
		return
	}
	defer file.Close()

	err = toml.NewEncoder(file).Encode(result.Plan)
	if err != nil {
		config.exitHandler.Error(err)
		return
	}
}
