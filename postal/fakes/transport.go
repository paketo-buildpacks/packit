package fakes

import (
	"io"
	"sync"
)

type Transport struct {
	DropCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Root string
			Uri  string
		}
		Returns struct {
			ReadCloser io.ReadCloser
			Error      error
		}
		Stub func(string, string) (io.ReadCloser, error)
	}
}

func (f *Transport) Drop(param1 string, param2 string) (io.ReadCloser, error) {
	f.DropCall.Lock()
	defer f.DropCall.Unlock()
	f.DropCall.CallCount++
	f.DropCall.Receives.Root = param1
	f.DropCall.Receives.Uri = param2
	if f.DropCall.Stub != nil {
		return f.DropCall.Stub(param1, param2)
	}
	return f.DropCall.Returns.ReadCloser, f.DropCall.Returns.Error
}
