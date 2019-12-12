package fakes

import "sync"

type DirectoryDuplicator struct {
	DuplicateCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			SourcePath string
			DestPath   string
		}
		Returns struct {
			Error error
		}
		Stub func(string, string) error
	}
}

func (f *DirectoryDuplicator) Duplicate(param1 string, param2 string) error {
	f.DuplicateCall.Lock()
	defer f.DuplicateCall.Unlock()
	f.DuplicateCall.CallCount++
	f.DuplicateCall.Receives.SourcePath = param1
	f.DuplicateCall.Receives.DestPath = param2
	if f.DuplicateCall.Stub != nil {
		return f.DuplicateCall.Stub(param1, param2)
	}
	return f.DuplicateCall.Returns.Error
}
