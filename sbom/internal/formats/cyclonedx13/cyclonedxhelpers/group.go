package cyclonedxhelpers

import "github.com/anchore/syft/syft/pkg"

// We must copy this helper in because it's not exported from
// syft/formats/common/cyclonedxhelpers
func encodeGroup(p pkg.Package) string {
	if hasMetadata(p) {
		if metadata, ok := p.Metadata.(pkg.JavaMetadata); ok && metadata.PomProperties != nil {
			return metadata.PomProperties.GroupID
		}
	}
	return ""
}
