package cyclonedxhelpers

import (
	"github.com/paketo-buildpacks/packit/v2/sbom/internal/formats/common"
	"github.com/paketo-buildpacks/packit/v2/sbom/internal/formats/cyclonedx13/cyclonedx"
)

// Relies on cycloneDX published structs
var (
	CycloneDXFields = common.RequiredTag("cyclonedx")
)

// We must copy this helper in because it's not exported from
// syft/formats/common/cyclonedxhelpers and it relies on our internal copy of cyclonedx 1.3
func encodeProperties(obj interface{}, prefix string) (out []cyclonedx.Property) {
	for _, p := range common.Sorted(common.Encode(obj, prefix, CycloneDXFields)) {
		out = append(out, cyclonedx.Property{
			Name:  p.Name,
			Value: p.Value,
		})
	}
	return
}
