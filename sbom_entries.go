package packit

import "io"

//go:generate faux --interface SBOMEntries --output fakes/sbom_entries.go
type SBOMEntries interface {
	Format() map[string]io.Reader
	IsEmpty() bool
}

type EmptySBOM struct{}

func (e EmptySBOM) Format() map[string]io.Reader {
	return make(map[string]io.Reader)
}

func (e EmptySBOM) IsEmpty() bool {
	return true
}
