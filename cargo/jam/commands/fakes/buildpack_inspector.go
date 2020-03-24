package fakes

import (
	"sync"

	"github.com/cloudfoundry/packit/cargo"
)

type BuildpackInspector struct {
	DependenciesCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			Dependencies []cargo.ConfigMetadataDependency
			Defaults     map[string]string
			Stacks       []string
			Err          error
		}
		Stub func(string) ([]cargo.ConfigMetadataDependency, map[string]string, []string, error)
	}
}

func (f *BuildpackInspector) Dependencies(param1 string) ([]cargo.ConfigMetadataDependency, map[string]string, []string, error) {
	f.DependenciesCall.Lock()
	defer f.DependenciesCall.Unlock()
	f.DependenciesCall.CallCount++
	f.DependenciesCall.Receives.Path = param1
	if f.DependenciesCall.Stub != nil {
		return f.DependenciesCall.Stub(param1)
	}
	return f.DependenciesCall.Returns.Dependencies, f.DependenciesCall.Returns.Defaults, f.DependenciesCall.Returns.Stacks, f.DependenciesCall.Returns.Err
}
