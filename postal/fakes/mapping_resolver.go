package fakes

import "sync"

type MappingResolver struct {
	FindDependencyMappingsCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			SHA256      string
			BindingPath string
		}
		Returns struct {
			String string
			Error  error
		}
		Stub func(string, string) (string, error)
	}
}

func (f *MappingResolver) FindDependencyMappings(param1 string, param2 string) (string, error) {
	f.FindDependencyMappingsCall.Lock()
	defer f.FindDependencyMappingsCall.Unlock()
	f.FindDependencyMappingsCall.CallCount++
	f.FindDependencyMappingsCall.Receives.SHA256 = param1
	f.FindDependencyMappingsCall.Receives.BindingPath = param2
	if f.FindDependencyMappingsCall.Stub != nil {
		return f.FindDependencyMappingsCall.Stub(param1, param2)
	}
	return f.FindDependencyMappingsCall.Returns.String, f.FindDependencyMappingsCall.Returns.Error
}
