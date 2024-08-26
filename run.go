package packit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2/internal"
)

// Run combines the invocation of both build and detect into a single entry
// point. Calling Run from an executable with a name matching "build" or
// "detect" will result in the matching DetectFunc or BuildFunc being called.
func Run(detect DetectFunc, build BuildFunc, options ...Option) {
	config := OptionConfig{
		exitHandler: internal.NewExitHandler(),
		args:        os.Args,
	}

	for _, option := range options {
		config = option(config)
	}

	phase := filepath.Base(config.args[0])

	switch phase {
	case "detect":
		Detect(detect, options...)
	case "build":
		Build(build, options...)
	default:
		config.exitHandler.Error(fmt.Errorf("failed to run buildpack: unknown lifecycle phase %q", phase))
	}

}

// RunExtension combines the invocation of both generate and detect into a single entry
// point. Calling Run from an executable with a name matching "generate" or
// "detect" will result in the matching DetectFunc or GenerateFunc being called.
func RunExtension(detect DetectFunc, generate GenerateFunc, options ...Option) {
	config := OptionConfig{
		exitHandler: internal.NewExitHandler(),
		args:        os.Args,
	}

	for _, option := range options {
		config = option(config)
	}

	phase := filepath.Base(config.args[0])

	switch phase {
	case "detect":
		Detect(detect, options...)
	case "generate":
		Generate(generate, options...)
	default:
		config.exitHandler.Error(fmt.Errorf("failed to run buildpack: unknown lifecycle phase %q", phase))
	}

}
