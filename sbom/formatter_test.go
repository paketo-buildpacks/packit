package sbom_test

import (
	"io"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testFormatter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		formatter sbom.Formatter
	)

	it.Before(func() {
		bom, err := sbom.Generate("testdata/")
		Expect(err).NotTo(HaveOccurred())

		formatter, err = bom.InFormats(sbom.CycloneDXFormat, sbom.SPDXFormat, sbom.SyftFormat)
		Expect(err).NotTo(HaveOccurred())
	})

	context("Format", func() {
		it("returns a copy of the original map", func() {
			// Assert that the first copy contains all of the right formats
			formats := formatter.Formats()
			Expect(formats).To(HaveLen(3))

			var extensions []string
			for _, format := range formats {
				extensions = append(extensions, format.Extension)
			}

			Expect(extensions).To(ConsistOf("cdx.json", "spdx.json", "syft.json"))

			for _, format := range formats {
				content, err := io.ReadAll(format.Content)
				Expect(err).NotTo(HaveOccurred())
				Expect(content).NotTo(BeEmpty())
			}

			// Assert that the second copy contains all of the right formats and that
			// the readers are repopulated
			formats = formatter.Formats()
			Expect(formats).To(HaveLen(3))

			extensions = []string{}
			for _, format := range formats {
				extensions = append(extensions, format.Extension)
			}

			Expect(extensions).To(ConsistOf("cdx.json", "spdx.json", "syft.json"))

			for _, format := range formats {
				content, err := io.ReadAll(format.Content)
				Expect(err).NotTo(HaveOccurred())
				Expect(content).NotTo(BeEmpty())
			}
		})
	})
}
