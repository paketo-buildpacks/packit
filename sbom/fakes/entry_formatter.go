package fakes

import (
	"io"
	"sync"

	"github.com/paketo-buildpacks/packit/sbom"
)

type EntryFormatter struct {
	FormatCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Format sbom.Format
		}
		Returns struct {
			Reader io.Reader
		}
		Stub func(sbom.Format) io.Reader
	}
	IsEmptyCall struct {
		mutex     sync.Mutex
		CallCount int
		Returns   struct {
			Bool bool
		}
		Stub func() bool
	}
}

func (f *EntryFormatter) Format(param1 sbom.Format) io.Reader {
	f.FormatCall.mutex.Lock()
	defer f.FormatCall.mutex.Unlock()
	f.FormatCall.CallCount++
	f.FormatCall.Receives.Format = param1
	if f.FormatCall.Stub != nil {
		return f.FormatCall.Stub(param1)
	}
	return f.FormatCall.Returns.Reader
}
func (f *EntryFormatter) IsEmpty() bool {
	f.IsEmptyCall.mutex.Lock()
	defer f.IsEmptyCall.mutex.Unlock()
	f.IsEmptyCall.CallCount++
	if f.IsEmptyCall.Stub != nil {
		return f.IsEmptyCall.Stub()
	}
	return f.IsEmptyCall.Returns.Bool
}
