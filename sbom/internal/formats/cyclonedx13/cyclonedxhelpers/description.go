package cyclonedxhelpers

import "github.com/anchore/syft/syft/pkg"

// We must copy this helper in because it's not exported from
// syft/formats/common/cyclonedxhelpers
func encodeDescription(p pkg.Package) string {
	if hasMetadata(p) {
		switch metadata := p.Metadata.(type) {
		case pkg.ApkMetadata:
			return metadata.Description
		case pkg.NpmPackageJSONMetadata:
			return metadata.Description
		}
	}
	return ""
}
