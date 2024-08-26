package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit/v2/servicebindings"
)

type BindingResolver struct {
	ResolveCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Typ         string
			Provider    string
			PlatformDir string
		}
		Returns struct {
			BindingSlice []servicebindings.Binding
			Error        error
		}
		Stub func(string, string, string) ([]servicebindings.Binding, error)
	}
}

func (f *BindingResolver) Resolve(param1 string, param2 string, param3 string) ([]servicebindings.Binding, error) {
	f.ResolveCall.mutex.Lock()
	defer f.ResolveCall.mutex.Unlock()
	f.ResolveCall.CallCount++
	f.ResolveCall.Receives.Typ = param1
	f.ResolveCall.Receives.Provider = param2
	f.ResolveCall.Receives.PlatformDir = param3
	if f.ResolveCall.Stub != nil {
		return f.ResolveCall.Stub(param1, param2, param3)
	}
	return f.ResolveCall.Returns.BindingSlice, f.ResolveCall.Returns.Error
}
