package sbom_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/anchore/syft/syft"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/sbom/internal/formats/syft2"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testFormattedReader(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		bom sbom.SBOM
	)

	it.Before(func() {
		var err error
		bom, err = sbom.Generate("testdata/")
		Expect(err).NotTo(HaveOccurred())
	})

	it("writes the SBOM in the default CycloneDX format", func() {
		buffer := bytes.NewBuffer(nil)
		_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.CycloneDXFormat))
		Expect(err).NotTo(HaveOccurred())

		format := syft.IdentifyFormat(buffer.Bytes())
		Expect(format.ID()).To(Equal(syft.CycloneDxJSONFormatID))

		var cdxOutput cdxOutput

		err = json.Unmarshal(buffer.Bytes(), &cdxOutput)
		Expect(err).NotTo(HaveOccurred(), buffer.String())

		Expect(cdxOutput.BOMFormat).To(Equal("CycloneDX"), buffer.String())
		Expect(cdxOutput.SpecVersion).To(Equal("1.3"), buffer.String())
		Expect(cdxOutput.SerialNumber).To(Equal(""), buffer.String())

		Expect(cdxOutput.Metadata.Timestamp).To(Equal(""), buffer.String())
		Expect(cdxOutput.Metadata.Component.Type).To(Equal("file"), buffer.String())
		Expect(cdxOutput.Metadata.Component.Type).To(Equal("file"), buffer.String())
		Expect(cdxOutput.Metadata.Component.Name).To(Equal("testdata/"), buffer.String())
		Expect(cdxOutput.Components[0].Name).To(Equal("collapse-white-space"), buffer.String())
		Expect(cdxOutput.Components[1].Name).To(Equal("end-of-stream"), buffer.String())
		Expect(cdxOutput.Components[2].Name).To(Equal("insert-css"), buffer.String())
		Expect(cdxOutput.Components[3].Name).To(Equal("once"), buffer.String())
		Expect(cdxOutput.Components[4].Name).To(Equal("pump"), buffer.String())
		Expect(cdxOutput.Components[5].Name).To(Equal("wrappy"), buffer.String())
	})

	it("writes the SBOM in the latest CycloneDX format (1.4)", func() {
		buffer := bytes.NewBuffer(nil)
		_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.Format(syft.CycloneDxJSONFormatID)))
		Expect(err).NotTo(HaveOccurred())

		format := syft.IdentifyFormat(buffer.Bytes())
		Expect(format.ID()).To(Equal(syft.CycloneDxJSONFormatID))

		var cdxOutput cdxOutput

		err = json.Unmarshal(buffer.Bytes(), &cdxOutput)
		Expect(err).NotTo(HaveOccurred(), buffer.String())

		Expect(cdxOutput.BOMFormat).To(Equal("CycloneDX"), buffer.String())
		Expect(cdxOutput.SpecVersion).To(Equal("1.4"), buffer.String())
		Expect(cdxOutput.SerialNumber).To(Equal(""), buffer.String())

		Expect(cdxOutput.Metadata.Timestamp).To(Equal(""), buffer.String())
		Expect(cdxOutput.Metadata.Component.Type).To(Equal("file"), buffer.String())
		Expect(cdxOutput.Metadata.Component.Name).To(Equal("testdata/"), buffer.String())
		Expect(cdxOutput.Components[0].Name).To(Equal("collapse-white-space"), buffer.String())
		Expect(cdxOutput.Components[1].Name).To(Equal("end-of-stream"), buffer.String())
		Expect(cdxOutput.Components[2].Name).To(Equal("insert-css"), buffer.String())
		Expect(cdxOutput.Components[3].Name).To(Equal("once"), buffer.String())
		Expect(cdxOutput.Components[4].Name).To(Equal("pump"), buffer.String())
		Expect(cdxOutput.Components[5].Name).To(Equal("wrappy"), buffer.String())
	})

	it("writes the SBOM in the default SPDX format", func() {
		buffer := bytes.NewBuffer(nil)
		_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.SPDXFormat))
		Expect(err).NotTo(HaveOccurred())

		var spdxOutput spdxOutput

		err = json.Unmarshal(buffer.Bytes(), &spdxOutput)
		Expect(err).NotTo(HaveOccurred(), buffer.String())

		Expect(spdxOutput.SPDXVersion).To(Equal("SPDX-2.2"), buffer.String())

		Expect(spdxOutput.Packages[0].Name).To(Equal("collapse-white-space"), buffer.String())
		Expect(spdxOutput.Packages[1].Name).To(Equal("end-of-stream"), buffer.String())
		Expect(spdxOutput.Packages[2].Name).To(Equal("insert-css"), buffer.String())
		Expect(spdxOutput.Packages[3].Name).To(Equal("once"), buffer.String())
		Expect(spdxOutput.Packages[4].Name).To(Equal("pump"), buffer.String())
		Expect(spdxOutput.Packages[5].Name).To(Equal("wrappy"), buffer.String())
	})

	it("writes the SBOM in the default syft format", func() {
		buffer := bytes.NewBuffer(nil)
		_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.SyftFormat))
		Expect(err).NotTo(HaveOccurred())

		var syftOutput syftOutput

		err = json.Unmarshal(buffer.Bytes(), &syftOutput)
		Expect(err).NotTo(HaveOccurred(), buffer.String())

		Expect(syftOutput.Schema.Version).To(Equal(`3.0.1`), buffer.String())

		Expect(syftOutput.Source.Type).To(Equal("directory"), buffer.String())
		Expect(syftOutput.Source.Target).To(Equal("testdata/"), buffer.String())
		Expect(syftOutput.Artifacts[0].Name).To(Equal("collapse-white-space"), buffer.String())
		Expect(syftOutput.Artifacts[1].Name).To(Equal("end-of-stream"), buffer.String())
		Expect(syftOutput.Artifacts[2].Name).To(Equal("insert-css"), buffer.String())
		Expect(syftOutput.Artifacts[3].Name).To(Equal("once"), buffer.String())
		Expect(syftOutput.Artifacts[4].Name).To(Equal("pump"), buffer.String())
		Expect(syftOutput.Artifacts[5].Name).To(Equal("wrappy"), buffer.String())
	})

	it("writes the SBOM in Syft 2.0.2 format", func() {
		buffer := bytes.NewBuffer(nil)
		_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.Format(syft2.ID)))
		Expect(err).NotTo(HaveOccurred())

		var syftOutput syftOutput

		err = json.Unmarshal(buffer.Bytes(), &syftOutput)
		Expect(err).NotTo(HaveOccurred(), buffer.String())

		Expect(syftOutput.Schema.Version).To(Equal("2.0.2"), buffer.String())

		Expect(syftOutput.Source.Type).To(Equal("directory"), buffer.String())
		Expect(syftOutput.Source.Target).To(Equal("testdata/"), buffer.String())
		Expect(syftOutput.Artifacts[0].Name).To(Equal("collapse-white-space"), buffer.String())
		Expect(syftOutput.Artifacts[1].Name).To(Equal("end-of-stream"), buffer.String())
		Expect(syftOutput.Artifacts[2].Name).To(Equal("insert-css"), buffer.String())
		Expect(syftOutput.Artifacts[3].Name).To(Equal("once"), buffer.String())
		Expect(syftOutput.Artifacts[4].Name).To(Equal("pump"), buffer.String())
		Expect(syftOutput.Artifacts[5].Name).To(Equal("wrappy"), buffer.String())
	})

	it("writes the SBOM in the latest Syft format (3.*)", func() {
		buffer := bytes.NewBuffer(nil)
		_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.Format(syft.JSONFormatID)))
		Expect(err).NotTo(HaveOccurred())

		var syftOutput syftOutput

		err = json.Unmarshal(buffer.Bytes(), &syftOutput)
		Expect(err).NotTo(HaveOccurred(), buffer.String())

		Expect(syftOutput.Schema.Version).To(MatchRegexp(`3\.\d+\.\d+`), buffer.String())

		Expect(syftOutput.Source.Type).To(Equal("directory"), buffer.String())
		Expect(syftOutput.Source.Target).To(Equal("testdata/"), buffer.String())
		Expect(syftOutput.Artifacts[0].Name).To(Equal("collapse-white-space"), buffer.String())
		Expect(syftOutput.Artifacts[1].Name).To(Equal("end-of-stream"), buffer.String())
		Expect(syftOutput.Artifacts[2].Name).To(Equal("insert-css"), buffer.String())
		Expect(syftOutput.Artifacts[3].Name).To(Equal("once"), buffer.String())
		Expect(syftOutput.Artifacts[4].Name).To(Equal("pump"), buffer.String())
		Expect(syftOutput.Artifacts[5].Name).To(Equal("wrappy"), buffer.String())
	})

	context("Read", func() {
		context("failure cases", func() {
			context("when the SBOM cannot be encoded to the given format", func() {
				it("returns an error", func() {
					formatter := sbom.NewFormattedReader(sbom.SBOM{}, sbom.Format("unknown-format"))
					_, err := formatter.Read(make([]byte, 10))
					Expect(err).To(MatchError("failed to format sbom: 'unknown-format' is not a valid SBOM format identifier"))
				})
			})
		})
	})
}
