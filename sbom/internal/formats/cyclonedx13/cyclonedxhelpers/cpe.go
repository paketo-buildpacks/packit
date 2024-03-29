package cyclonedxhelpers

import (
	"github.com/anchore/syft/syft/cpe"
	"github.com/anchore/syft/syft/pkg"
	"github.com/paketo-buildpacks/packit/v2/sbom/internal/formats/cyclonedx13/cyclonedx"
)

// Relies on cycloneDX published structs
// We must copy this helper in because it's not exported from
// syft/formats/common/cyclonedxhelpers
func encodeSingleCPE(p pkg.Package) string {
	// Since the CPEs in a package are sorted by specificity
	// we can extract the first CPE as the one to output in cyclonedx
	if len(p.CPEs) > 0 {
		return cpe.String(p.CPEs[0])
	}
	return ""
}

func encodeCPEs(p pkg.Package) (out []cyclonedx.Property) {
	for i, c := range p.CPEs {
		// first CPE is "most specific" and already encoded as the component CPE
		if i == 0 {
			continue
		}
		out = append(out, cyclonedx.Property{
			Name:  "syft:cpe23",
			Value: cpe.String(c),
		})
	}
	return
}
