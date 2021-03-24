package fakes

import "sync"

type ProjectPathParser struct {
	GetCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Path       string
			EnvVarName string
		}
		Returns struct {
			ProjectPath string
			Err         error
		}
		Stub func(string, string) (string, error)
	}
}

func (f *ProjectPathParser) Get(param1 string, param2 string) (string, error) {
	f.GetCall.Lock()
	defer f.GetCall.Unlock()
	f.GetCall.CallCount++
	f.GetCall.Receives.Path = param1
	f.GetCall.Receives.EnvVarName = param2
	if f.GetCall.Stub != nil {
		return f.GetCall.Stub(param1, param2)
	}
	return f.GetCall.Returns.ProjectPath, f.GetCall.Returns.Err
}
