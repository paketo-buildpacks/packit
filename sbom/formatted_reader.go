package sbom

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/sbom"
)

// FormattedReader outputs the SBoM in a specified format.
type FormattedReader struct {
	m      sync.Mutex
	sbom   SBOM
	format Format
	reader io.Reader
}

// NewFormattedReader creates an instance of FormattedReader given an SBOM and
// Format.
func NewFormattedReader(s SBOM, f Format) *FormattedReader {
	return &FormattedReader{sbom: s, format: f}
}

// Read implements the io.Reader interface to output the contents of the
// formatted SBoM.
func (f *FormattedReader) Read(b []byte) (int, error) {
	f.m.Lock()
	defer f.m.Unlock()

	if f.reader == nil {
		var id sbom.FormatID
		switch f.format {
		case CycloneDXFormat:
			id = syft.CycloneDxJSONFormatID
		case SPDXFormat:
			id = syft.SPDXJSONFormatID
		case SyftFormat:
			id = syft.JSONFormatID
		default:
			return 0, fmt.Errorf("failed to format sbom: unsupported format %q", f.format)
		}

		output, err := syft.Encode(f.sbom.syft, syft.FormatByID(id))
		if err != nil {
			return 0, fmt.Errorf("failed to format sbom: %w", err)
		}

		f.reader = bytes.NewBuffer(output)
	}

	return f.reader.Read(b)
}
