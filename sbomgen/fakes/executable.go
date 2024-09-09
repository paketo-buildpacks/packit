package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit/v2/pexec"
)

type Executable struct {
	ExecuteCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Execution pexec.Execution
		}
		Returns struct {
			Err error
		}
		Stub func(pexec.Execution) error
	}
}

func (f *Executable) Execute(param1 pexec.Execution) error {
	f.ExecuteCall.mutex.Lock()
	defer f.ExecuteCall.mutex.Unlock()
	f.ExecuteCall.CallCount++
	f.ExecuteCall.Receives.Execution = param1
	if f.ExecuteCall.Stub != nil {
		return f.ExecuteCall.Stub(param1)
	}
	return f.ExecuteCall.Returns.Err
}
