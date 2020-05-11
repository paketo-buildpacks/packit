package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit/pexec"
)

type Executable struct {
	ExecuteCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Execution pexec.Execution
		}
		Returns struct {
			Error error
		}
		Stub func(pexec.Execution) error
	}
}

func (f *Executable) Execute(param1 pexec.Execution) error {
	f.ExecuteCall.Lock()
	defer f.ExecuteCall.Unlock()
	f.ExecuteCall.CallCount++
	f.ExecuteCall.Receives.Execution = param1
	if f.ExecuteCall.Stub != nil {
		return f.ExecuteCall.Stub(param1)
	}
	return f.ExecuteCall.Returns.Error
}
