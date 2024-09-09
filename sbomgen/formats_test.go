package sbomgen_test

import (
	"testing"

	"github.com/paketo-buildpacks/packit/v2/sbomgen"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testFormats(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect
	var f sbomgen.Format

	context("Formats", func() {
		context("no version param", func() {
			it("gets the right mediatype extension and version", func() {
				f = sbomgen.CycloneDXFormat
				ext, err := f.Extension()
				Expect(err).NotTo(HaveOccurred())
				Expect(ext).To(Equal("cdx.json"))
				Expect(f.VersionParam()).To(Equal(""))
			})
		})

		context("with version param", func() {
			it("gets the right mediatype extension and version", func() {
				f = sbomgen.SPDXFormat + ";version=9.8.7"
				ext, err := f.Extension()
				Expect(err).NotTo(HaveOccurred())
				Expect(ext).To(Equal("spdx.json"))
				Expect(f.VersionParam()).To(Equal("9.8.7"))
			})
			context("Syft mediatype with version returns empty", func() {
				it("returns error", func() {
					f = sbomgen.SyftFormat + ";version=9.8.7"
					ext, err := f.Extension()
					Expect(err).To(MatchError(ContainSubstring("Unknown mediatype application/vnd.syft+json;version=9.8.7")))
					Expect(ext).To(Equal(""))
				})
			})
		})
	})
}
