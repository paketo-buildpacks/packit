package sbom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/sbom"
)

// FormattedReader outputs the SBoM in a specified format.
type FormattedReader struct {
	m           sync.Mutex
	sbom        SBOM
	rawFormatID string
	format      sbom.Format
	reader      io.Reader
}

// NewFormattedReader creates an instance of FormattedReader given an SBOM and
// Format.
func NewFormattedReader(s SBOM, f Format) *FormattedReader {
	// For backward compatibility, caller can pass f as a format ID like
	// "cyclonedx-1.3-json" or as a media type like
	// 'application/vnd.cyclonedx+json'
	sbomFormat, err := sbomFormatByID(sbom.FormatID(f))
	if err != nil {
		sbomFormat, err = sbomFormatByMediaType(string(f))
		if err != nil {
			// Defer throwing an error until Read() is called
			return &FormattedReader{sbom: s, rawFormatID: string(f), format: nil}
		}
	}
	return &FormattedReader{sbom: s, rawFormatID: string(sbomFormat.ID()), format: sbomFormat}
}

// Read implements the io.Reader interface to output the contents of the
// formatted SBoM.
func (f *FormattedReader) Read(b []byte) (int, error) {
	f.m.Lock()
	defer f.m.Unlock()

	if f.reader == nil {
		if f.format == nil {
			return 0, fmt.Errorf("failed to format sbom: '%s' is not a valid SBOM format identifier", f.rawFormatID)
		}

		output, err := syft.Encode(f.sbom.syft, f.format)
		if err != nil {
			// not tested
			return 0, fmt.Errorf("failed to format sbom: %w", err)
		}

		// Makes CycloneDX SBOM more reproducible, see
		// https://github.com/paketo-buildpacks/packit/issues/367 for more details.
		if f.format.ID() == "cyclonedx-1.3-json" || f.format.ID() == "cyclonedx-1-json" {
			var cycloneDXOutput map[string]interface{}
			err = json.Unmarshal(output, &cycloneDXOutput)
			if err != nil {
				return 0, fmt.Errorf("failed to modify CycloneDX SBOM for reproducibility: %w", err)
			}
			for k := range cycloneDXOutput {
				if k == "metadata" {
					metadata := cycloneDXOutput[k].(map[string]interface{})
					delete(metadata, "timestamp")
					cycloneDXOutput[k] = metadata
				}
				if k == "serialNumber" {
					delete(cycloneDXOutput, k)
				}
			}
			output, err = json.Marshal(cycloneDXOutput)
			if err != nil {
				return 0, fmt.Errorf("failed to modify CycloneDX SBOM for reproducibility: %w", err)
			}
		}

		f.reader = bytes.NewBuffer(output)
	}

	return f.reader.Read(b)
}
