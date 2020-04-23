package packit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/cloudfoundry/packit/internal"
)

// BuildContext provides the contextual details that are made available by the
// buildpack lifecycle during the build phase. This context is populated by the
// Build function and passed to BuildFunc during execution.
type BuildContext struct {
	// BuildpackInfo includes the details of the buildpack parsed from the
	// buildpack.toml included in the buildpack contents.
	BuildpackInfo BuildpackInfo

	// CNBPath is the absolute path location of the buildpack contents.
	// This path is useful for finding the buildpack.toml or any other
	// files included in the buildpack.
	CNBPath string

	// Layers provides access to layers managed by the buildpack. It can be used
	// to create new layers or retrieve cached layers from previous builds.
	Layers Layers

	// Plan includes the BuildpackPlan provided by the lifecycle as specified in
	// the specification:
	// https://github.com/buildpacks/spec/blob/master/buildpack.md#buildpack-plan-toml.
	Plan BuildpackPlan

	// Stack is the value of the chosen stack. This value is populated from the
	// $CNB_STACK_ID environment variable.
	Stack string

	// WorkingDir is the location of the application source code as provided by
	// the lifecycle.
	WorkingDir string
}

// BuildFunc is the definition of a callback that can be invoked when the Build
// function is executed. Buildpack authors should implement a BuildFunc that
// performs the specific build phase operations for a buildpack.
type BuildFunc func(BuildContext) (BuildResult, error)

// BuildResult allows buildpack authors to indicate the result of the build
// phase for a given buildpack. This result, returned in a BuildFunc callback,
// will be parsed and persisted by the Build function and returned to the
// lifecycle at the end of the build phase execution.
type BuildResult struct {
	// Plan is the set of refinements to the Buildpack Plan that were performed
	// during the build phase.
	Plan BuildpackPlan

	// Layers is a list of layers that will be persisted by the lifecycle at the
	// end of the build phase. Layers not included in this list will not be made
	// available to the lifecycle.
	Layers []Layer

	// Processes is a list of processes that will be returned to the lifecycle to
	// be executed during the launch phase.
	Processes []Process
}

// Process represents a process to be run during the launch phase as described
// in the specification:
// https://github.com/buildpacks/spec/blob/master/buildpack.md#launch. The
// fields of the process are describe in the specification of the launch.toml
// file:
// https://github.com/buildpacks/spec/blob/master/buildpack.md#launchtoml-toml.
type Process struct {
	// Type is an identifier to describe the type of process to be executed, eg.
	// "web".
	Type string `toml:"type"`

	// Command is the start command to be executed at launch.
	Command string `toml:"command"`

	// Args is a list of arguments to be passed to the command at launch.
	Args []string `toml:"args"`

	// Direct indicates whether the process should bypass the shell when invoked.
	Direct bool `toml:"direct"`
}

// BuildpackInfo is a representation of the basic information for a buildpack
// provided in its buildpack.toml file as described in the specification:
// https://github.com/buildpacks/spec/blob/master/buildpack.md#buildpacktoml-toml.
type BuildpackInfo struct {
	// ID is the identifier specified in the `buildpack.id` field of the buildpack.toml.
	ID string `toml:"id"`

	// Name is the identifier specified in the `buildpack.name` field of the buildpack.toml.
	Name string `toml:"name"`

	// Version is the identifier specified in the `buildpack.version` field of the buildpack.toml.
	Version string `toml:"version"`
}

// BuildpackPlan is a representation of the buildpack plan provided by the
// lifecycle and defined in the specification:
// https://github.com/buildpacks/spec/blob/master/buildpack.md#buildpack-plan-toml.
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

	// Version if the version contraint that defines what would be an acceptable
	// dependency provided by the buildpack.
	Version string `toml:"version"`

	// Metadata is an unspecified field allowing buildpacks to communicate extra
	// details about their requirement. Examples of this type of metadata might
	// include details about what source was used to decide the Version
	// constraint for a requirement.
	Metadata map[string]interface{} `toml:"metadata"`
}

// Build is an implementation of the build phase according to the Cloud Native
// Buildpacks specification. Calling this function with a BuildFunc will
// perform the build phase process.
func Build(f BuildFunc, options ...Option) {
	config := OptionConfig{
		exitHandler: internal.NewExitHandler(),
		args:        os.Args,
		tomlWriter:  internal.NewTOMLWriter(),
		envWriter:   internal.NewEnvironmentWriter(),
	}

	for _, option := range options {
		config = option(config)
	}

	var (
		layersPath = config.args[1]
		planPath   = config.args[3]
	)

	pwd, err := os.Getwd()
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

	var plan BuildpackPlan
	_, err = toml.DecodeFile(planPath, &plan)
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

	cnbPath := filepath.Clean(strings.TrimSuffix(config.args[0], filepath.Join("bin", "build")))

	var buildpackInfo struct {
		Buildpack BuildpackInfo `toml:"buildpack"`
	}
	_, err = toml.DecodeFile(filepath.Join(cnbPath, "buildpack.toml"), &buildpackInfo)
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

	result, err := f(BuildContext{
		CNBPath:    cnbPath,
		Stack:      os.Getenv("CNB_STACK_ID"),
		WorkingDir: pwd,
		Plan:       plan,
		Layers: Layers{
			Path: layersPath,
		},
		BuildpackInfo: buildpackInfo.Buildpack,
	})
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

	err = config.tomlWriter.Write(planPath, result.Plan)
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

	layerTomls, err := filepath.Glob(filepath.Join(layersPath, "*.toml"))
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

	for _, file := range layerTomls {
		if filepath.Base(file) != "launch.toml" && filepath.Base(file) != "store.toml" {
			err = os.Remove(file)
			if err != nil {
				config.exitHandler.Error(fmt.Errorf("failed to remove layer toml: %w", err))
				return
			}
		}
	}

	for _, layer := range result.Layers {
		err = config.tomlWriter.Write(filepath.Join(layersPath, fmt.Sprintf("%s.toml", layer.Name)), layer)
		if err != nil {
			config.exitHandler.Error(err)
			return
		}

		err = config.envWriter.Write(filepath.Join(layer.Path, "env"), layer.SharedEnv)
		if err != nil {
			config.exitHandler.Error(err)
			return
		}

		err = config.envWriter.Write(filepath.Join(layer.Path, "env.launch"), layer.LaunchEnv)
		if err != nil {
			config.exitHandler.Error(err)
			return
		}

		err = config.envWriter.Write(filepath.Join(layer.Path, "env.build"), layer.BuildEnv)
		if err != nil {
			config.exitHandler.Error(err)
			return
		}
	}

	if len(result.Processes) > 0 {
		var launch struct {
			Processes []Process `toml:"processes"`
		}
		launch.Processes = result.Processes

		err = config.tomlWriter.Write(filepath.Join(layersPath, "launch.toml"), launch)
		if err != nil {
			config.exitHandler.Error(err)
			return
		}
	}
}
