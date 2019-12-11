package fakes

import "sync"

type PrePackager struct {
	ExecuteCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Path    string
			RootDir string
		}
		Returns struct {
			Error error
		}
		Stub func(string, string) error
	}
}

func (f *PrePackager) Execute(param1 string, param2 string) error {
	f.ExecuteCall.Lock()
	defer f.ExecuteCall.Unlock()
	f.ExecuteCall.CallCount++
	f.ExecuteCall.Receives.Path = param1
	f.ExecuteCall.Receives.RootDir = param2
	if f.ExecuteCall.Stub != nil {
		return f.ExecuteCall.Stub(param1, param2)
	}
	return f.ExecuteCall.Returns.Error
}
