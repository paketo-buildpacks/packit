package sbom

import (
	"bytes"
	"io"
	"sync"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/format"
)

type SyftFormatter struct {
	m      sync.Mutex
	sbom   SBOM
	reader io.Reader
}

func NewSyftFormatter(s SBOM) *SyftFormatter {
	return &SyftFormatter{sbom: s}
}

func (f *SyftFormatter) Read(b []byte) (int, error) {
	f.m.Lock()
	defer f.m.Unlock()

	if f.reader == nil {
		output, err := syft.Encode(f.sbom.syft, format.JSONOption)
		if err != nil {
			panic(err)
		}

		f.reader = bytes.NewBuffer(output)
	}

	return f.reader.Read(b)
}
