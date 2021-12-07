package packit

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/packit/v2/internal"
)

// BuildFunc is the definition of a callback that can be invoked when the Build
// function is executed. Buildpack authors should implement a BuildFunc that
// performs the specific build phase operations for a buildpack.
type BuildFunc func(BuildContext) (BuildResult, error)

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

	// Platform includes the platform context according to the specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#build
	Platform Platform

	// Layers provides access to layers managed by the buildpack. It can be used
	// to create new layers or retrieve cached layers from previous builds.
	Layers Layers

	// Plan includes the BuildpackPlan provided by the lifecycle as specified in
	// the specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#buildpack-plan-toml.
	Plan BuildpackPlan

	// Stack is the value of the chosen stack. This value is populated from the
	// $CNB_STACK_ID environment variable.
	Stack string

	// WorkingDir is the location of the application source code as provided by
	// the lifecycle.
	WorkingDir string
}

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

	// Launch is the metadata that will be persisted as launch.toml according to
	// the buildpack lifecycle specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#launchtoml-toml
	Launch LaunchMetadata

	// Build is the metadata that will be persisted as build.toml according to
	// the buildpack lifecycle specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#buildtoml-toml
	Build BuildMetadata
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
		fileWriter:  internal.NewFileWriter(),
	}

	for _, option := range options {
		config = option(config)
	}

	var (
		layersPath   = config.args[1]
		platformPath = config.args[2]
		planPath     = config.args[3]
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

	cnbPath, ok := os.LookupEnv("CNB_BUILDPACK_DIR")
	if !ok {
		cnbPath = filepath.Clean(strings.TrimSuffix(config.args[0], filepath.Join("bin", "build")))
	}

	var buildpackInfo struct {
		APIVersion string        `toml:"api"`
		Buildpack  BuildpackInfo `toml:"buildpack"`
	}

	_, err = toml.DecodeFile(filepath.Join(cnbPath, "buildpack.toml"), &buildpackInfo)
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

	apiV05, _ := semver.NewVersion("0.5")
	apiV06, _ := semver.NewVersion("0.6")
	apiVersion, err := semver.NewVersion(buildpackInfo.APIVersion)
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

	result, err := f(BuildContext{
		CNBPath: cnbPath,
		Platform: Platform{
			Path: platformPath,
		},
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

	if len(result.Plan.Entries) > 0 {
		if apiVersion.GreaterThan(apiV05) || apiVersion.Equal(apiV05) {
			config.exitHandler.Error(errors.New("buildpack plan is read only since Buildpack API v0.5"))
			return
		}

		err = config.tomlWriter.Write(planPath, result.Plan)
		if err != nil {
			config.exitHandler.Error(err)
			return
		}
	}

	layerTomls, err := filepath.Glob(filepath.Join(layersPath, "*.toml"))
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

	if apiVersion.LessThan(apiV06) {
		for _, file := range layerTomls {
			if filepath.Base(file) != "launch.toml" && filepath.Base(file) != "store.toml" && filepath.Base(file) != "build.toml" {
				err = os.Remove(file)
				if err != nil {
					config.exitHandler.Error(fmt.Errorf("failed to remove layer toml: %w", err))
					return
				}
			}
		}
	}

	for _, layer := range result.Layers {
		err = config.tomlWriter.Write(filepath.Join(layersPath, fmt.Sprintf("%s.toml", layer.Name)), formattedLayer{layer, apiVersion})
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

		for process, processEnv := range layer.ProcessLaunchEnv {
			err = config.envWriter.Write(filepath.Join(layer.Path, "env.launch", process), processEnv)
			if err != nil {
				config.exitHandler.Error(err)
				return
			}
		}

		if layer.SBOM != nil {
			if apiVersion.GreaterThan(apiV06) {
				for _, format := range layer.SBOM.Formats() {
					err = config.fileWriter.Write(filepath.Join(layersPath, fmt.Sprintf("%s.sbom.%s", layer.Name, format.Extension)), format.Content)
					if err != nil {
						config.exitHandler.Error(err)
						return
					}
				}
			} else {
				config.exitHandler.Error(fmt.Errorf("%s.sbom.* output is only supported with Buildpack API v0.7 or higher", layer.Name))
				return
			}
		}
	}

	if !result.Launch.isEmpty() {
		if apiVersion.LessThan(apiV05) && len(result.Launch.BOM) > 0 {
			config.exitHandler.Error(errors.New("BOM entries in launch.toml is only supported with Buildpack API v0.5 or higher"))
			return
		}

		type label struct {
			Key   string `toml:"key"`
			Value string `toml:"value"`
		}

		var launch struct {
			Processes []Process  `toml:"processes"`
			Slices    []Slice    `toml:"slices"`
			Labels    []label    `toml:"labels"`
			BOM       []BOMEntry `toml:"bom"`
		}

		launch.Processes = result.Launch.Processes
		if apiVersion.LessThan(apiV06) {
			for _, process := range launch.Processes {
				if process.Default {
					config.exitHandler.Error(errors.New("processes can only be marked as default with Buildpack API v0.6 or higher"))
					return
				}
			}
		}

		launch.Slices = result.Launch.Slices
		launch.BOM = result.Launch.BOM
		if len(result.Launch.Labels) > 0 {
			launch.Labels = []label{}
			for k, v := range result.Launch.Labels {
				launch.Labels = append(launch.Labels, label{Key: k, Value: v})
			}

			sort.Slice(launch.Labels, func(i, j int) bool {
				return launch.Labels[i].Key < launch.Labels[j].Key
			})
		}

		err = config.tomlWriter.Write(filepath.Join(layersPath, "launch.toml"), launch)
		if err != nil {
			config.exitHandler.Error(err)
			return
		}

		if result.Launch.SBOM != nil {
			if apiVersion.GreaterThan(apiV06) {
				for _, format := range result.Launch.SBOM.Formats() {
					err = config.fileWriter.Write(filepath.Join(layersPath, fmt.Sprintf("launch.sbom.%s", format.Extension)), format.Content)
					if err != nil {
						config.exitHandler.Error(err)
						return
					}
				}
			} else {
				config.exitHandler.Error(fmt.Errorf("launch.sbom.* output is only supported with Buildpack API v0.7 or higher"))
				return
			}
		}
	}

	if !result.Build.isEmpty() {
		if apiVersion.LessThan(apiV05) {
			config.exitHandler.Error(fmt.Errorf("build.toml is only supported with Buildpack API v0.5 or higher"))
			return
		}

		if result.Build.SBOM != nil {
			if apiVersion.GreaterThan(apiV06) {
				for _, format := range result.Build.SBOM.Formats() {
					err = config.fileWriter.Write(filepath.Join(layersPath, fmt.Sprintf("build.sbom.%s", format.Extension)), format.Content)
					if err != nil {
						config.exitHandler.Error(err)
						return
					}
				}
			} else {
				config.exitHandler.Error(fmt.Errorf("build.sbom.* output is only supported with Buildpack API v0.7 or higher"))
				return
			}
		}
		err = config.tomlWriter.Write(filepath.Join(layersPath, "build.toml"), result.Build)
		if err != nil {
			config.exitHandler.Error(err)
			return
		}
	}
}
