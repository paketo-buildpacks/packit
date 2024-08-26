package spdx22

import (
	"io"

	"github.com/anchore/syft/syft/sbom"
)

const ID sbom.FormatID = "spdx-2-json"

func Format() sbom.Format {
	return sbom.NewFormat(
		"2.2",
		encoder,
		func(input io.Reader) (*sbom.SBOM, error) { return nil, nil },
		func(input io.Reader) error { return nil },
		ID,
	)
}
