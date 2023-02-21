package syft2

import (
	"io"

	"github.com/anchore/syft/syft/sbom"
)

const ID sbom.FormatID = "syft-2.0-json"
const JSONSchemaVersion string = "2.0.2"

func Format() sbom.Format {
	return sbom.NewFormat(
		"2.0.2",
		encoder,
		func(input io.Reader) (*sbom.SBOM, error) { return nil, nil },
		func(input io.Reader) error { return nil },
		ID,
	)
}
