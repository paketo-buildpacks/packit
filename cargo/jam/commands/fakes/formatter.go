package fakes

import (
	"sync"

	"github.com/cloudfoundry/packit/cargo"
)

type Formatter struct {
	MarkdownCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Dependencies    []cargo.ConfigMetadataDependency
			DefaultVersions map[string]string
		}
		Stub func([]cargo.ConfigMetadataDependency, map[string]string)
	}
}

func (f *Formatter) Markdown(param1 []cargo.ConfigMetadataDependency, param2 map[string]string) {
	f.MarkdownCall.Lock()
	defer f.MarkdownCall.Unlock()
	f.MarkdownCall.CallCount++
	f.MarkdownCall.Receives.Dependencies = param1
	f.MarkdownCall.Receives.DefaultVersions = param2
	if f.MarkdownCall.Stub != nil {
		f.MarkdownCall.Stub(param1, param2)
	}
}
