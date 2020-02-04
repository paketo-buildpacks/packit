package packit

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/cloudfoundry/packit/internal"
)

var Fail = internal.Fail

type BuildPlanProvision struct {
	Name string `toml:"name"`
}

type BuildPlanRequirement struct {
	Name     string      `toml:"name"`
	Version  string      `toml:"version"`
	Metadata interface{} `toml:"metadata"`
}

type BuildPlan struct {
	Provides []BuildPlanProvision   `toml:"provides"`
	Requires []BuildPlanRequirement `toml:"requires"`
}

type DetectContext struct {
	WorkingDir    string
	BuildpackInfo BuildpackInfo
}

type DetectResult struct {
	Plan BuildPlan
}

type DetectFunc func(DetectContext) (DetectResult, error)

func Detect(f DetectFunc, options ...Option) {
	config := Config{
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

	cnbPath := filepath.Clean(strings.TrimSuffix(config.args[0], filepath.Join("bin", "detect")))

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
		BuildpackInfo: buildpackInfo.Buildpack,
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
