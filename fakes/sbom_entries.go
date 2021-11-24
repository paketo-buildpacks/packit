package fakes

import (
	"io"
	"sync"
)

type SBOMEntries struct {
	FormatCall struct {
		mutex     sync.Mutex
		CallCount int
		Returns   struct {
			MapStringIoReader map[string]io.Reader
		}
		Stub func() map[string]io.Reader
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

func (f *SBOMEntries) Format() map[string]io.Reader {
	f.FormatCall.mutex.Lock()
	defer f.FormatCall.mutex.Unlock()
	f.FormatCall.CallCount++
	if f.FormatCall.Stub != nil {
		return f.FormatCall.Stub()
	}
	return f.FormatCall.Returns.MapStringIoReader
}
func (f *SBOMEntries) IsEmpty() bool {
	f.IsEmptyCall.mutex.Lock()
	defer f.IsEmptyCall.mutex.Unlock()
	f.IsEmptyCall.CallCount++
	if f.IsEmptyCall.Stub != nil {
		return f.IsEmptyCall.Stub()
	}
	return f.IsEmptyCall.Returns.Bool
}
