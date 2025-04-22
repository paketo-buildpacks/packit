package sbom

import (
	"fmt"
	"mime"

	"github.com/anchore/syft/syft/format"
	syftsbom "github.com/anchore/syft/syft/sbom"
)

var encoderCollection = format.NewEncoderCollection(format.Encoders()...)

var cyclonedxFormats map[string]syftsbom.FormatID = map[string]syftsbom.FormatID{
	"default": CycloneDX13,
	"1.3":     CycloneDX13,
	"1.4":     CycloneDX14,
}

var spdxFormats map[string]syftsbom.FormatID = map[string]syftsbom.FormatID{
	"default": SPDX22,
	"2.2":     SPDX22,
}

// An experimental type added to support more SBOM formats
// It extends the Syft sbom.Format interface
type sbomFormat struct {
	syftsbom.FormatEncoder
}

func newSBOMFormat(format syftsbom.FormatEncoder) sbomFormat {
	return sbomFormat{
		FormatEncoder: format,
	}
}

func (f sbomFormat) Extension() string {
	switch f.ID() {
	case CycloneLatest, CycloneDX13, CycloneDX14:
		return "cdx.json"
	case SPDXLatest, SPDX22:
		return "spdx.json"
	case SyftLatest:
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
	var selected syftsbom.FormatID
	switch baseType {
	case CycloneDXFormat:
		selected = cyclonedxFormats[version]
	case SPDXFormat:
		selected = spdxFormats[version]
	case SyftFormat:
		selected = SyftLatest
	default:
		return sbomFormat{}, fmt.Errorf("unsupported SBOM format: '%s'", mediaType)
	}

	if selected == syftsbom.FormatID("") {
		return sbomFormat{}, fmt.Errorf("version '%s' is not supported for SBOM format '%s'", version, baseType)
	}
	return sbomFormatByID(selected)
}

func sbomFormatByID(id syftsbom.FormatID) (sbomFormat, error) {
	format := encoderCollection.GetByString(id.String())
	if format == nil {
		return sbomFormat{}, fmt.Errorf("'%s' is not a valid SBOM format identifier", id)
	}
	return newSBOMFormat(format), nil
}
