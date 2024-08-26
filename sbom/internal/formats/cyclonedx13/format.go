package cyclonedx13

import (
	"io"

	"github.com/anchore/syft/syft/sbom"
)

// TODO: Decide version granularity of IDs
const ID sbom.FormatID = "cyclonedx-1.3-json"

func Format() sbom.Format {
	return sbom.NewFormat(
		sbom.AnyVersion,
		encoder,
		func(input io.Reader) (*sbom.SBOM, error) { return nil, nil },
		func(input io.Reader) error { return nil },
		ID,
	)
}
