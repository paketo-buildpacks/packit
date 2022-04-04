package cyclonedxhelpers

import (
	"github.com/anchore/syft/syft/pkg"
	"github.com/paketo-buildpacks/packit/v2/sbom/internal/formats/cyclonedx13/cyclonedx"
)

// Relies on cycloneDX published structs
func encodeComponent(p pkg.Package) cyclonedx.Component {
	props := encodeProperties(p, "syft:package")
	props = append(props, encodeCPEs(p)...)
	locations := p.Locations.ToSlice()
	if len(locations) > 0 {
		props = append(props, encodeProperties(locations, "syft:location")...)
	}
	if hasMetadata(p) {
		props = append(props, encodeProperties(p.Metadata, "syft:metadata")...)
	}

	var properties *[]cyclonedx.Property
	if len(props) > 0 {
		properties = &props
	}

	return cyclonedx.Component{
		Type:               cyclonedx.ComponentTypeLibrary,
		Name:               p.Name,
		Group:              encodeGroup(p),
		Version:            p.Version,
		PackageURL:         p.PURL,
		Licenses:           encodeLicenses(p),
		CPE:                encodeSingleCPE(p),
		Author:             encodeAuthor(p),
		Publisher:          encodePublisher(p),
		Description:        encodeDescription(p),
		ExternalReferences: encodeExternalReferences(p),
		Properties:         properties,
	}
}

func hasMetadata(p pkg.Package) bool {
	return p.Metadata != nil
}
