package fakes

import (
	"sync"

	"github.com/cloudfoundry/packit/cargo"
)

type ConfigParser struct {
	ParseCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			Config cargo.Config
			Error  error
		}
		Stub func(string) (cargo.Config, error)
	}
}

func (f *ConfigParser) Parse(param1 string) (cargo.Config, error) {
	f.ParseCall.Lock()
	defer f.ParseCall.Unlock()
	f.ParseCall.CallCount++
	f.ParseCall.Receives.Path = param1
	if f.ParseCall.Stub != nil {
		return f.ParseCall.Stub(param1)
	}
	return f.ParseCall.Returns.Config, f.ParseCall.Returns.Error
}
