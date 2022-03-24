package cyclonedxhelpers

import (
	"github.com/paketo-buildpacks/packit/v2/sbom/internal/formats/common"
	"github.com/paketo-buildpacks/packit/v2/sbom/internal/formats/cyclonedx13/cyclonedx"
)

// Relies on cycloneDX published structs
var (
	CycloneDXFields = common.RequiredTag("cyclonedx")
)

func encodeProperties(obj interface{}, prefix string) (out []cyclonedx.Property) {
	for _, p := range common.Sorted(common.Encode(obj, prefix, CycloneDXFields)) {
		out = append(out, cyclonedx.Property{
			Name:  p.Name,
			Value: p.Value,
		})
	}
	return
}
