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
	format sbom.FormatID
	reader io.Reader
}

// NewFormattedReader creates an instance of FormattedReader given an SBOM and
// Format.
func NewFormattedReader(s SBOM, f Format) *FormattedReader {
	// type conversion to maintain backward compatibility of constructor
	return &FormattedReader{sbom: s, format: sbom.FormatID(f)}
}

// Read implements the io.Reader interface to output the contents of the
// formatted SBoM.
func (f *FormattedReader) Read(b []byte) (int, error) {
	f.m.Lock()
	defer f.m.Unlock()

	if f.reader == nil {
		// To maintain backward compatibility for users passing the exported consts
		// CycloneDXFormat, SPDXFormat, and SyftFormat into this function
		// wrap f.format in ensureFormatID
		format, err := sbomFormatByID(ensureFormatID(string(f.format)))
		if err != nil {
			return 0, fmt.Errorf("failed to format sbom: %w", err)
		}

		output, err := syft.Encode(f.sbom.syft, format)
		if err != nil {
			// not tested
			return 0, fmt.Errorf("failed to format sbom: %w", err)
		}

		f.reader = bytes.NewBuffer(output)
	}

	return f.reader.Read(b)
}

// Converts from exported strings CycloneDXFormat, SPDXFormat, and SyftFormat
// (whose values are actually media types) into corresponding FormatIDs
func ensureFormatID(mediaType string) sbom.FormatID {
	switch mediaType {
	case CycloneDXFormat:
		return syft.CycloneDxJSONFormatID
	case SPDXFormat:
		return syft.SPDXJSONFormatID
	case SyftFormat:
		return syft.JSONFormatID
	default:
		return sbom.FormatID(mediaType)
	}
}
