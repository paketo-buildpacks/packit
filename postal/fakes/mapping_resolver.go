package fakes

import "sync"

type MappingResolver struct {
	FindDependencyMappingCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Checksum    string
			PlatformDir string
		}
		Returns struct {
			String string
			Error  error
		}
		Stub func(string, string) (string, error)
	}
}

func (f *MappingResolver) FindDependencyMapping(param1 string, param2 string) (string, error) {
	f.FindDependencyMappingCall.mutex.Lock()
	defer f.FindDependencyMappingCall.mutex.Unlock()
	f.FindDependencyMappingCall.CallCount++
	f.FindDependencyMappingCall.Receives.Checksum = param1
	f.FindDependencyMappingCall.Receives.PlatformDir = param2
	if f.FindDependencyMappingCall.Stub != nil {
		return f.FindDependencyMappingCall.Stub(param1, param2)
	}
	return f.FindDependencyMappingCall.Returns.String, f.FindDependencyMappingCall.Returns.Error
}
