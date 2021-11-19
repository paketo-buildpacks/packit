package sbom

const (
	CycloneDXFormat = "application/vnd.cyclonedx+json"
	SPDXFormat      = "application/spdx+json"
	SyftFormat      = "application/vnd.syft+json"
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
