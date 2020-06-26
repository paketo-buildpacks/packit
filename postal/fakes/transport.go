package fakes

import (
	"io"
	"net/http"
	"sync"
)

type Transport struct {
	DropCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Root   string
			Uri    string
			Header http.Header
		}
		Returns struct {
			ReadCloser io.ReadCloser
			Error      error
		}
		Stub func(string, string, http.Header) (io.ReadCloser, error)
	}
}

func (f *Transport) Drop(param1 string, param2 string, param3 http.Header) (io.ReadCloser, error) {
	f.DropCall.Lock()
	defer f.DropCall.Unlock()
	f.DropCall.CallCount++
	f.DropCall.Receives.Root = param1
	f.DropCall.Receives.Uri = param2
	f.DropCall.Receives.Header = param3
	if f.DropCall.Stub != nil {
		return f.DropCall.Stub(param1, param2, param3)
	}
	return f.DropCall.Returns.ReadCloser, f.DropCall.Returns.Error
}
