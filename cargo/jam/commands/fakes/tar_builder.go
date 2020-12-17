package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit/cargo/jam/internal"
)

type TarBuilder struct {
	BuildCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Path  string
			Files []internal.File
		}
		Returns struct {
			Error error
		}
		Stub func(string, []internal.File) error
	}
}

func (f *TarBuilder) Build(param1 string, param2 []internal.File) error {
	f.BuildCall.Lock()
	defer f.BuildCall.Unlock()
	f.BuildCall.CallCount++
	f.BuildCall.Receives.Path = param1
	f.BuildCall.Receives.Files = param2
	if f.BuildCall.Stub != nil {
		return f.BuildCall.Stub(param1, param2)
	}
	return f.BuildCall.Returns.Error
}
