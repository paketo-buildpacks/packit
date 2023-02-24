package syft301

import (
	"io"

	"github.com/anchore/syft/syft/sbom"
)

const ID sbom.FormatID = "syft-3.0.1-json"
const JSONSchemaVersion string = "3.0.1"

func Format() sbom.Format {
	return sbom.NewFormat(
		"3.0.1",
		encoder,
		func(input io.Reader) (*sbom.SBOM, error) { return nil, nil },
		func(input io.Reader) error { return nil },
		ID,
	)
}
