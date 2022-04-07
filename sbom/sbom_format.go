package sbom

import (
	"fmt"
	"mime"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/sbom"
	"github.com/paketo-buildpacks/packit/v2/sbom/internal/formats/cyclonedx13"
	"github.com/paketo-buildpacks/packit/v2/sbom/internal/formats/syft2"
	"github.com/paketo-buildpacks/packit/v2/sbom/internal/formats/syft301"
)

// TODO: refactor the version lookup part
var syftFormats map[string]sbom.FormatID = map[string]sbom.FormatID{
	"default": syft301.ID,
	"3.0.1":   syft301.ID,
	"2.0.2":   syft2.ID,
}

var cyclonedxFormats map[string]sbom.FormatID = map[string]sbom.FormatID{
	"default": cyclonedx13.ID,
	"1.4":     syft.CycloneDxJSONFormatID,
	"1.3":     cyclonedx13.ID,
}

var spdxFormats map[string]sbom.FormatID = map[string]sbom.FormatID{
	"default": syft.SPDXJSONFormatID,
	"2.2":     syft.SPDXJSONFormatID,
}

var additionalFormats []sbomFormat

func init() {
	additionalFormats = []sbomFormat{
		newSBOMFormat(cyclonedx13.Format()),
		newSBOMFormat(syft2.Format()),
		newSBOMFormat(syft301.Format()),
	}
}

// An experimental type added to support more SBOM formats
// It extends the Syft sbom.Format interface
type sbomFormat struct {
	sbom.Format
}

func newSBOMFormat(format sbom.Format) sbomFormat {
	return sbomFormat{
		Format: format,
	}
}

func (f sbomFormat) Extension() string {
	switch f.ID() {
	case syft.CycloneDxJSONFormatID, cyclonedx13.ID:
		return "cdx.json"
	case syft.SPDXJSONFormatID:
		return "spdx.json"
	case syft.JSONFormatID, syft2.ID, syft301.ID:
		return "syft.json"
	default:
		return ""
	}
}

func sbomFormatByMediaType(mediaType string) (sbomFormat, error) {
	baseType, params, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return sbomFormat{}, fmt.Errorf("failed to parse SBOM media type: %w", err)
	}
	// TODO: semver version parsing?
	version, ok := params["version"]
	if !ok {
		version = "default"
	}
	var selected sbom.FormatID
	switch baseType {
	case CycloneDXFormat:
		selected = cyclonedxFormats[version]
	case SPDXFormat:
		selected = spdxFormats[version]
	case SyftFormat:
		selected = syftFormats[version]
	default:
		return sbomFormat{}, fmt.Errorf("unsupported SBOM format: '%s'", mediaType)
	}

	if selected == sbom.FormatID("") {
		return sbomFormat{}, fmt.Errorf("version '%s' is not supported for SBOM format '%s'", version, baseType)
	}
	return sbomFormatByID(selected)
}

func sbomFormatByID(id sbom.FormatID) (sbomFormat, error) {
	for _, f := range additionalFormats {
		if f.ID() == id {
			return f, nil
		}
	}
	format := syft.FormatByID(id)
	if format == nil {
		return sbomFormat{}, fmt.Errorf("'%s' is not a valid SBOM format identifier", id)
	}
	return newSBOMFormat(format), nil
}
