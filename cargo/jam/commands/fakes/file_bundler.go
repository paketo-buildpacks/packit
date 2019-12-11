package fakes

import (
	"sync"

	"github.com/cloudfoundry/packit/cargo"
)

type FileBundler struct {
	BundleCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Path   string
			Files  []string
			Config cargo.Config
		}
		Returns struct {
			FileSlice []cargo.File
			Error     error
		}
		Stub func(string, []string, cargo.Config) ([]cargo.File, error)
	}
}

func (f *FileBundler) Bundle(param1 string, param2 []string, param3 cargo.Config) ([]cargo.File, error) {
	f.BundleCall.Lock()
	defer f.BundleCall.Unlock()
	f.BundleCall.CallCount++
	f.BundleCall.Receives.Path = param1
	f.BundleCall.Receives.Files = param2
	f.BundleCall.Receives.Config = param3
	if f.BundleCall.Stub != nil {
		return f.BundleCall.Stub(param1, param2, param3)
	}
	return f.BundleCall.Returns.FileSlice, f.BundleCall.Returns.Error
}
