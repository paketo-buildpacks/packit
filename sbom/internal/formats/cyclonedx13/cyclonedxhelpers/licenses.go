package cyclonedxhelpers

import (
	"github.com/anchore/syft/syft/pkg"
	"github.com/paketo-buildpacks/packit/v2/sbom/internal/formats/cyclonedx13/cyclonedx"
	"github.com/paketo-buildpacks/packit/v2/sbom/internal/spdxlicense"
)

// Relies on cycloneDX published structs
func encodeLicenses(p pkg.Package) *cyclonedx.Licenses {
	lc := cyclonedx.Licenses{}
	for _, licenseName := range p.Licenses {
		if value, exists := spdxlicense.ID(licenseName); exists {
			lc = append(lc, cyclonedx.LicenseChoice{
				License: &cyclonedx.License{
					ID: value,
				},
			})
		}
	}
	if len(lc) > 0 {
		return &lc
	}
	return nil
}

func decodeLicenses(c *cyclonedx.Component) (out []string) {
	if c.Licenses != nil {
		for _, l := range *c.Licenses {
			out = append(out, l.License.ID)
		}
	}
	return
}
