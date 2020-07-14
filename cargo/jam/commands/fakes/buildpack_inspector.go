package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit/cargo"
)

type BuildpackInspector struct {
	DependenciesCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			ConfigSlice []cargo.Config
			Error       error
		}
		Stub func(string) ([]cargo.Config, error)
	}
}

func (f *BuildpackInspector) Dependencies(param1 string) ([]cargo.Config, error) {
	f.DependenciesCall.Lock()
	defer f.DependenciesCall.Unlock()
	f.DependenciesCall.CallCount++
	f.DependenciesCall.Receives.Path = param1
	if f.DependenciesCall.Stub != nil {
		return f.DependenciesCall.Stub(param1)
	}
	return f.DependenciesCall.Returns.ConfigSlice, f.DependenciesCall.Returns.Error
}
