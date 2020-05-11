package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit/cargo"
)

type Formatter struct {
	MarkdownCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Dependencies []cargo.ConfigMetadataDependency
			Defaults     map[string]string
			Stacks       []string
		}
		Stub func([]cargo.ConfigMetadataDependency, map[string]string, []string)
	}
}

func (f *Formatter) Markdown(param1 []cargo.ConfigMetadataDependency, param2 map[string]string, param3 []string) {
	f.MarkdownCall.Lock()
	defer f.MarkdownCall.Unlock()
	f.MarkdownCall.CallCount++
	f.MarkdownCall.Receives.Dependencies = param1
	f.MarkdownCall.Receives.Defaults = param2
	f.MarkdownCall.Receives.Stacks = param3
	if f.MarkdownCall.Stub != nil {
		f.MarkdownCall.Stub(param1, param2, param3)
	}
}
