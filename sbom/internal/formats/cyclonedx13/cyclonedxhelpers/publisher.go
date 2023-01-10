package cyclonedxhelpers

import (
	"github.com/anchore/syft/syft/pkg"
)

// We must copy this helper in because it's not exported from
// syft/formats/common/cyclonedxhelpers
func encodePublisher(p pkg.Package) string {
	if hasMetadata(p) {
		switch metadata := p.Metadata.(type) {
		case pkg.ApkMetadata:
			return metadata.Maintainer
		case pkg.RpmMetadata:
			return metadata.Vendor
		case pkg.DpkgMetadata:
			return metadata.Maintainer
		}
	}
	return ""
}
