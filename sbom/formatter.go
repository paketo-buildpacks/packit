package sbom

import (
	"github.com/anchore/syft/syft/sbom"
	"github.com/paketo-buildpacks/packit/v2"
)

// Formatter implements the packit.SBOMFormatter interface.
type Formatter struct {
	sbom      SBOM
	formatIDs []sbom.FormatID
}

// Formats returns a list of packit.SBOMFormat instances.
func (f Formatter) Formats() []packit.SBOMFormat {
	var formats []packit.SBOMFormat
	for _, id := range f.formatIDs {
		// ignore error here; FormattedReader validates SBOM format before Read()
		format, _ := sbomFormatByID(id)
		formats = append(formats, packit.SBOMFormat{
			Extension: format.Extension(),
			// type conversion here to maintain backward compatibility of NewFormattedReader
			Content: NewFormattedReader(f.sbom, Format(id)),
		})
	}

	return formats
}
