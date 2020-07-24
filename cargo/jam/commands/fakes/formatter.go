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
			ConfigSlice []cargo.Config
		}
		Stub func([]cargo.Config)
	}
}

func (f *Formatter) Markdown(param1 []cargo.Config) {
	f.MarkdownCall.Lock()
	defer f.MarkdownCall.Unlock()
	f.MarkdownCall.CallCount++
	f.MarkdownCall.Receives.ConfigSlice = param1
	if f.MarkdownCall.Stub != nil {
		f.MarkdownCall.Stub(param1)
	}
}
