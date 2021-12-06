package fakes

import "sync"

type ExitHandler struct {
	ErrorCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Error error
		}
		Stub func(error)
	}
}

func (f *ExitHandler) Error(param1 error) {
	f.ErrorCall.mutex.Lock()
	defer f.ErrorCall.mutex.Unlock()
	f.ErrorCall.CallCount++
	f.ErrorCall.Receives.Error = param1
	if f.ErrorCall.Stub != nil {
		f.ErrorCall.Stub(param1)
	}
}
