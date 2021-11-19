package packit

import "io"

type SBOMEntries map[string]io.Reader

func (e SBOMEntries) Set(path string, reader io.Reader) {
	e[path] = reader
}
