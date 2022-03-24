package cyclonedx13

import (
	"io"

	"github.com/anchore/syft/syft/sbom"
)

// TODO: Decide version granularity of IDs
const ID sbom.FormatID = "cyclonedx-1.3-json"

var dummyDecoder func(io.Reader) (*sbom.SBOM, error) = func(input io.Reader) (*sbom.SBOM, error) {
	return nil, nil
}

var dummyValidator func(io.Reader) error = func(input io.Reader) error {
	return nil
}

func Format() sbom.Format {
	return sbom.NewFormat(
		ID,
		encoder,
		dummyDecoder,
		dummyValidator,
	)
}
