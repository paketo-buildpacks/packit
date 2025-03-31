package sbom

import (
	syftsbom "github.com/anchore/syft/syft/sbom"
)

const (
	SyftFormat      = "application/vnd.syft+json"
	CycloneDXFormat = "application/vnd.cyclonedx+json"
	SPDXFormat      = "application/spdx+json"

	SyftLatest = syftsbom.FormatID("syft-json")

	CycloneLatest = syftsbom.FormatID("cyclonedx-json")
	CycloneDX13   = syftsbom.FormatID("cyclonedx-json@1.3")
	CycloneDX14   = syftsbom.FormatID("cyclonedx-json@1.4")

	SPDXLatest = syftsbom.FormatID("spdx-json")
	SPDX22     = syftsbom.FormatID("spdx-json@2.2")
)

// Format is the type declaration for the supported SBoM output formats.
type Format string

// Extension outputs the expected file extension for a given Format.
func (f Format) Extension() string {
	switch f {
	case CycloneDXFormat:
		return "cdx.json"
	case SPDXFormat:
		return "spdx.json"
	case SyftFormat:
		return "syft.json"
	default:
		return ""
	}
}
