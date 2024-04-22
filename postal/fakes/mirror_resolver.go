package fakes

import "sync"

type MirrorResolver struct {
	FindDependencyMirrorCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Uri         string
			PlatformDir string
		}
		Returns struct {
			String string
			Error  error
		}
		Stub func(string, string) (string, error)
	}
}

func (f *MirrorResolver) FindDependencyMirror(param1 string, param2 string) (string, error) {
	f.FindDependencyMirrorCall.mutex.Lock()
	defer f.FindDependencyMirrorCall.mutex.Unlock()
	f.FindDependencyMirrorCall.CallCount++
	f.FindDependencyMirrorCall.Receives.Uri = param1
	f.FindDependencyMirrorCall.Receives.PlatformDir = param2
	if f.FindDependencyMirrorCall.Stub != nil {
		return f.FindDependencyMirrorCall.Stub(param1, param2)
	}
	return f.FindDependencyMirrorCall.Returns.String, f.FindDependencyMirrorCall.Returns.Error
}
