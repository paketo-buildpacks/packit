package sbom

import "github.com/paketo-buildpacks/packit/v2"

// Formatter implements the packit.SBOMFormatter interface.
type Formatter struct {
	sbom    SBOM
	formats []Format
}

// Formats returns a list of packit.SBOMFormat instances.
func (f Formatter) Formats() []packit.SBOMFormat {
	var formats []packit.SBOMFormat
	for _, format := range f.formats {
		formats = append(formats, packit.SBOMFormat{
			Extension: format.Extension(),
			Content:   NewFormattedReader(f.sbom, format),
		})
	}

	return formats
}
