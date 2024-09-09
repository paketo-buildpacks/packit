package sbomgen

import (
	"fmt"
	"mime"
	"strings"
)

const (
	CycloneDXFormat = "application/vnd.cyclonedx+json"
	SPDXFormat      = "application/spdx+json"
	SyftFormat      = "application/vnd.syft+json"
)

// Format is the type declaration for the supported SBoM output formats.
type Format string

// Extension outputs the expected file extension for a given Format.
// packit allows CycloneDX and SPDX mediatypes to have an optional
// version suffix. e.g. "application/vnd.cyclonedx+json;version=1.4"
// The version suffix is not allowed for the syft mediatype as the
// syft tooling does not support providing a version for this mediatype.
func (f Format) Extension() (string, error) {
	switch {
	case strings.HasPrefix(string(f), CycloneDXFormat):
		return "cdx.json", nil
	case strings.HasPrefix(string(f), SPDXFormat):
		return "spdx.json", nil
	case f == SyftFormat:
		return "syft.json", nil
	default:
		return "", fmt.Errorf("Unknown mediatype %s", f)
	}
}

// Extracts optional version. This usually derives from the "sbom-formats"
// field used by packit-based buildpacks (@packit.SBOMFormats). e.g.
// "application/vnd.cyclonedx+json;version=1.4" -> "1.4" See
// github.com/paketo-buildpacks/packit/issues/302
func (f Format) VersionParam() (string, error) {
	_, params, err := mime.ParseMediaType(string(f))
	if err != nil {
		return "", fmt.Errorf("failed to parse SBOM mediatype. Expected <mediatype>[;version=<ver>], Got %s: %w", f, err)
	}

	version, ok := params["version"]
	if !ok {
		return "", nil
	}
	return version, nil
}
