package sbom

import (
	"bytes"
	"io"
	"sync"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/format"
)

type CycloneDXFormatter struct {
	m      sync.Mutex
	sbom   SBOM
	reader io.Reader
}

func NewCycloneDXFormatter(s SBOM) *CycloneDXFormatter {
	return &CycloneDXFormatter{sbom: s}
}

func (f *CycloneDXFormatter) Read(b []byte) (int, error) {
	f.m.Lock()
	defer f.m.Unlock()

	if f.reader == nil {
		output, err := syft.Encode(f.sbom.syft, format.CycloneDxJSONOption)
		if err != nil {
			panic(err)
		}

		f.reader = bytes.NewBuffer(output)
	}

	return f.reader.Read(b)
}
