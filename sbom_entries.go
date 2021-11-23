package packit

import "io"

type SBOMEntries interface {
	Format() (map[string]io.Reader, error)
	IsEmpty() bool
}

type EmptySBOM struct{}

func (e EmptySBOM) Format() (map[string]io.Reader, error) {
	return make(map[string]io.Reader), nil
}

func (e EmptySBOM) IsEmpty() bool {
	return true
}
