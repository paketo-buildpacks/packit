package sbom

import (
	"bytes"
	"io"
	"sync"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/format"
)

type SPDXFormatter struct {
	m      sync.Mutex
	sbom   SBOM
	reader io.Reader
}

func NewSPDXFormatter(s SBOM) *SPDXFormatter {
	return &SPDXFormatter{sbom: s}
}

func (f *SPDXFormatter) Read(b []byte) (int, error) {
	f.m.Lock()
	defer f.m.Unlock()

	if f.reader == nil {
		output, err := syft.Encode(f.sbom.syft, format.SPDXJSONOption)
		if err != nil {
			panic(err)
		}

		f.reader = bytes.NewBuffer(output)
	}

	return f.reader.Read(b)
}
