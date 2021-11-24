package sbom_test

import (
	"io"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/packit/sbom"
	"github.com/paketo-buildpacks/packit/sbom/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testEntries(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect        = NewWithT(t).Expect
		reader        io.Reader
		sbomFormatter *fakes.EntryFormatter
	)

	it.Before(func() {
		sbomFormatter = &fakes.EntryFormatter{}
		reader = strings.NewReader(`{"some-key": "some-value"}`)
		sbomFormatter.FormatCall.Returns.Reader = reader
	})

	context("Format", func() {
		context("when no formats have been added", func() {
			var entries sbom.Entries
			it.Before(func() {
				entries = sbom.NewEntries(sbomFormatter)
			})
			it("returns an empty initialized map", func() {
				result := entries.Format()
				Expect(result).NotTo(BeNil())
				Expect(result).To(BeEmpty())
			})
		})

		context("when formats have been added", func() {
			var entries sbom.Entries
			it.Before(func() {
				entries = sbom.NewEntries(sbomFormatter)
				entries.AddFormat(sbom.SPDXFormat)
				entries.AddFormat(sbom.CycloneDXFormat)
			})
			it("returns a map of file extensions to formatted content readers", func() {
				result := entries.Format()
				Expect(result).To(Equal(map[string]io.Reader{
					sbom.SPDXFormat.Extension():      reader,
					sbom.CycloneDXFormat.Extension(): reader,
				}))
			})
		})
	})

	context("IsEmpty", func() {
		context("when the SBOM content is empty", func() {
			var entries sbom.Entries
			it.Before(func() {
				entries = sbom.NewEntries(sbomFormatter)
				sbomFormatter.IsEmptyCall.Returns.Bool = true
			})
			it("returns true", func() {
				Expect(entries.IsEmpty()).To(BeTrue())
			})
		})
		context("when the SBOM content is not empty", func() {
			var entries sbom.Entries
			it.Before(func() {
				entries = sbom.NewEntries(sbomFormatter)
				sbomFormatter.IsEmptyCall.Returns.Bool = false
			})
			it("returns false", func() {
				Expect(entries.IsEmpty()).To(BeFalse())
			})
		})
	})
	context("GetContent", func() {
		it("returns the formatted SBOM", func() {
			entries := sbom.NewEntries(sbomFormatter)
			Expect(entries.GetContent(sbom.SPDXFormat)).To(Equal(reader))
			Expect(sbomFormatter.FormatCall.Receives.Format).To(Equal(sbom.SPDXFormat))
		})
	})
}
