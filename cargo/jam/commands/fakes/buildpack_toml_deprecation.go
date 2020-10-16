package fakes

import "sync"

type BuildpackTOMLDeprecation struct {
	WarnDeprecatedFieldsCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			Error error
		}
		Stub func(string) error
	}
}

func (f *BuildpackTOMLDeprecation) WarnDeprecatedFields(param1 string) error {
	f.WarnDeprecatedFieldsCall.Lock()
	defer f.WarnDeprecatedFieldsCall.Unlock()
	f.WarnDeprecatedFieldsCall.CallCount++
	f.WarnDeprecatedFieldsCall.Receives.Path = param1
	if f.WarnDeprecatedFieldsCall.Stub != nil {
		return f.WarnDeprecatedFieldsCall.Stub(param1)
	}
	return f.WarnDeprecatedFieldsCall.Returns.Error
}
