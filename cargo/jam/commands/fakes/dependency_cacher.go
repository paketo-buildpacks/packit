package fakes

import (
	"sync"

	"github.com/cloudfoundry/packit/cargo"
)

type DependencyCacher struct {
	CacheCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Root         string
			Dependencies []cargo.ConfigMetadataDependency
		}
		Returns struct {
			ConfigMetadataDependencySlice []cargo.ConfigMetadataDependency
			Error                         error
		}
		Stub func(string, []cargo.ConfigMetadataDependency) ([]cargo.ConfigMetadataDependency, error)
	}
}

func (f *DependencyCacher) Cache(param1 string, param2 []cargo.ConfigMetadataDependency) ([]cargo.ConfigMetadataDependency, error) {
	f.CacheCall.Lock()
	defer f.CacheCall.Unlock()
	f.CacheCall.CallCount++
	f.CacheCall.Receives.Root = param1
	f.CacheCall.Receives.Dependencies = param2
	if f.CacheCall.Stub != nil {
		return f.CacheCall.Stub(param1, param2)
	}
	return f.CacheCall.Returns.ConfigMetadataDependencySlice, f.CacheCall.Returns.Error
}
