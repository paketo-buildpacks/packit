package spdx22

import (
	"io"

	"github.com/anchore/syft/syft/sbom"
)

const ID sbom.FormatID = "spdx-2-json"

// Decoder not implemented because it's not needed for buildpacks' SBOM generation
var dummyDecoder func(io.Reader) (*sbom.SBOM, error) = func(input io.Reader) (*sbom.SBOM, error) {
	return nil, nil
}

// Validator not implemented because it's not needed for buildpacks' SBOM generation
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
