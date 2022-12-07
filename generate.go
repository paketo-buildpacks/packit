package packit

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/packit/v2/internal"
)

// GenerateFunc is the definition of a callback that can be invoked when the Generate
// function is executed. Extension authors should implement a GenerateFunc that
// performs the specific generate phase operations for that extension.
type GenerateFunc func(GenerateContext) (GenerateResult, error)

// GenerateContext provides the contextual details that are made available by the
// extension lifecycle during the generate phase. This context is populated by the
// Generate function and passed to GenerateFunc during execution.
type GenerateContext struct {
	// Info includes the details of the buildpack parsed from the
	// extension.toml included in the extension contents.
	Info Info

	// CNBPath is the absolute path location of the buildpack contents.
	// This path is useful for finding the buildpack.toml or any other
	// files included in the buildpack.
	CNBPath string

	// Platform includes the platform context according to the specification:
	// https://github.com/buildpacks/spec/blob/main/buildpack.md#build
	Platform Platform

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

// GenerateResult allows extension authors to indicate the result of the generate
// phase for a given extension. This result, returned in a GenerateFunc callback,
// will be parsed and persisted by the Generate function and returned to the
// lifecycle at the end of the generate phase execution.
type GenerateResult struct {
	// ExtendConfig contains the config of an extension
	ExtendConfig ExtendConfig

	// BuildDockerfile the Dockerfile to define the build image
	BuildDockerfile io.Reader
	// RunDockerfile the Dockerfile to define the run image
	RunDockerfile io.Reader
}

type ExtendConfig struct {
	Build ExtendImageConfig `toml:"build"`
}

type ExtendImageConfig struct {
	Args []ExtendImageConfigArg `toml:"args"`
}

type ExtendImageConfigArg struct {
	Name  string `toml:"name"`
	Value string `toml:"value"`
}

// Generate is an implementation of the generate phase according to the Cloud Native
// Buildpacks specification. Calling this function with a GenerateFunc will
// perform the generate phase process of an extension.
func Generate(f GenerateFunc, options ...Option) {
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

	pwd, err := os.Getwd()
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

	planPath := os.Getenv("CNB_BP_PLAN_PATH")

	var plan BuildpackPlan
	_, err = toml.DecodeFile(planPath, &plan)
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

	cnbPath := os.Getenv("CNB_EXTENSION_DIR")
	outputPath := os.Getenv("CNB_OUTPUT_DIR")
	platformPath := os.Getenv("CNB_PLATFORM_DIR")

	var info struct {
		APIVersion string `toml:"api"`
		Info       Info   `toml:"extension"`
	}

	extensionTOML := filepath.Join(cnbPath, "extension.toml")
	_, err = toml.DecodeFile(extensionTOML, &info)
	if err != nil {
		config.exitHandler.Error(fmt.Errorf("could not parse %q: %w", extensionTOML, err))
		return
	}

	result, err := f(GenerateContext{
		CNBPath: cnbPath,
		Platform: Platform{
			Path: platformPath,
		},
		Stack:      os.Getenv("CNB_STACK_ID"),
		WorkingDir: pwd,
		Plan:       plan,
		Info:       info.Info,
	})
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

	if result.BuildDockerfile != nil {
		err = config.fileWriter.Write(filepath.Join(outputPath, "build.Dockerfile"), result.BuildDockerfile)
		if err != nil {
			config.exitHandler.Error(err)
			return
		}
	}
	if result.RunDockerfile != nil {
		err = config.fileWriter.Write(filepath.Join(outputPath, "run.Dockerfile"), result.RunDockerfile)
		if err != nil {
			config.exitHandler.Error(err)
			return
		}
	}

	err = config.tomlWriter.Write(filepath.Join(outputPath, "extend-config.toml"), result.ExtendConfig)
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

}
