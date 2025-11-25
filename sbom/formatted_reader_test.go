package sbom_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"

	"github.com/anchore/syft/syft/format"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testFormattedReader(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		bom sbom.SBOM
		err error
	)

	it.Before(func() {
		bom, err = sbom.Generate("testdata/")
		Expect(err).NotTo(HaveOccurred())
	})

	it("writes the SBOM in the 1.3 CycloneDX format", func() {
		buffer := bytes.NewBuffer(nil)
		_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.CycloneDXFormat))
		Expect(err).NotTo(HaveOccurred())

		format, version := format.Identify(bytes.NewBuffer(buffer.Bytes()))
		Expect(format).To(Equal(sbom.CycloneLatest))
		Expect(version).To(Equal("1.3"))

		// Ensures pretty printing
		Expect(buffer.String()).To(ContainSubstring(`{
  "$schema": "http://cyclonedx.org/schema/bom-1.3.schema.json",
  "bomFormat": "CycloneDX",
  "components": [
    {`))

		var cdxOutput cdxOutput

		err = json.Unmarshal(buffer.Bytes(), &cdxOutput)
		Expect(err).NotTo(HaveOccurred(), buffer.String())

		Expect(cdxOutput.BOMFormat).To(Equal("CycloneDX"), buffer.String())
		Expect(cdxOutput.SpecVersion).To(Equal("1.3"), buffer.String())
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

		rerunBuffer := bytes.NewBuffer(nil)
		_, err = io.Copy(rerunBuffer, sbom.NewFormattedReader(bom, sbom.CycloneDXFormat))
		Expect(err).NotTo(HaveOccurred())
		Expect(rerunBuffer.String()).To(Equal(buffer.String()))
	})

	it("writes the SBOM in the 1.4 CycloneDX format", func() {
		buffer := bytes.NewBuffer(nil)
		_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.Format(sbom.CycloneDX14)))
		Expect(err).NotTo(HaveOccurred())

		// Ensures pretty printing
		Expect(buffer.String()).To(ContainSubstring(`{
  "$schema": "http://cyclonedx.org/schema/bom-1.4.schema.json",
  "bomFormat": "CycloneDX",
  "components": [
`))

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

		rerunBuffer := bytes.NewBuffer(nil)
		_, err = io.Copy(rerunBuffer, sbom.NewFormattedReader(bom, sbom.Format(sbom.CycloneDX14)))
		Expect(err).NotTo(HaveOccurred())
		Expect(rerunBuffer.String()).To(Equal(buffer.String()))
	})

	context("writes the SBOM in latest SPDX format, with fields replaced for reproducibility", func() {
		it("produces an SBOM", func() {
			buffer := bytes.NewBuffer(nil)
			_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.SPDXFormat))
			Expect(err).NotTo(HaveOccurred())

			format, version := format.Identify(bytes.NewBuffer(buffer.Bytes()))
			Expect(format).To(Equal(sbom.SPDXLatest))
			Expect(version).To(Equal("2.2"))

			// Ensures pretty printing
			Expect(buffer.String()).To(ContainSubstring(`{
 "SPDXID": "SPDXRef-DOCUMENT",
 "creationInfo": {`))

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

			// Ensure documentNamespace and creationInfo.created have reproducible values
			// The DocumentNamespace includes the sha1 uuid of the data, which may differ if running locally versus in github actions
			Expect(spdxOutput.DocumentNamespace).To(MatchRegexp(`^https://paketo\.io/packit/dir/testdata-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`), buffer.String())
			Expect(spdxOutput.CreationInfo.Created).To(BeZero(), buffer.String())

			rerunBuffer := bytes.NewBuffer(nil)
			_, err = io.Copy(rerunBuffer, sbom.NewFormattedReader(bom, sbom.SPDXFormat))
			Expect(err).NotTo(HaveOccurred())
			Expect(rerunBuffer.String()).To(Equal(buffer.String()))
		})

		context("when SOURCE_DATE_EPOCH is set", func() {
			var original string

			it.Before(func() {
				original = os.Getenv("SOURCE_DATE_EPOCH")
				Expect(os.Setenv("SOURCE_DATE_EPOCH", "1659551872")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Setenv("SOURCE_DATE_EPOCH", original)).To(Succeed())
			})

			context("when the timestamp is valid", func() {
				it.Before(func() {
					Expect(os.Setenv("SOURCE_DATE_EPOCH", "1659551872")).To(Succeed())
				})

				it("produces an SBOM with the given timestamp", func() {
					buffer := bytes.NewBuffer(nil)
					_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.SPDXFormat))
					Expect(err).NotTo(HaveOccurred())

					var spdxOutput spdxOutput

					err = json.Unmarshal(buffer.Bytes(), &spdxOutput)
					Expect(err).NotTo(HaveOccurred(), buffer.String())

					format, version := format.Identify(bytes.NewBuffer(buffer.Bytes()))
					Expect(format).To(Equal(sbom.SPDXLatest))
					Expect(version).To(Equal("2.2"))

					Expect(spdxOutput.SPDXVersion).To(Equal("SPDX-2.2"), buffer.String())

					Expect(spdxOutput.Packages[0].Name).To(Equal("collapse-white-space"), buffer.String())
					Expect(spdxOutput.Packages[1].Name).To(Equal("end-of-stream"), buffer.String())
					Expect(spdxOutput.Packages[2].Name).To(Equal("insert-css"), buffer.String())
					Expect(spdxOutput.Packages[3].Name).To(Equal("once"), buffer.String())
					Expect(spdxOutput.Packages[4].Name).To(Equal("pump"), buffer.String())
					Expect(spdxOutput.Packages[5].Name).To(Equal("wrappy"), buffer.String())

					// Ensure documentNamespace and creationInfo.created have reproducible values
					// The DocumentNamespace includes the sha1 uuid of the data, which may differ if running locally versus in github actions
					Expect(spdxOutput.DocumentNamespace).To(MatchRegexp(`^https://paketo\.io/packit/dir/testdata-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`), buffer.String())
					Expect(spdxOutput.CreationInfo.Created).To(Equal(time.Unix(1659551872, 0).UTC()), buffer.String())

					rerunBuffer := bytes.NewBuffer(nil)
					_, err = io.Copy(rerunBuffer, sbom.NewFormattedReader(bom, sbom.SPDXFormat))
					Expect(err).NotTo(HaveOccurred())
					Expect(rerunBuffer.String()).To(Equal(buffer.String()))
				})

				context("failure cases", func() {
					context("when the timestamp is not valid", func() {
						it.Before(func() {
							Expect(os.Setenv("SOURCE_DATE_EPOCH", "not-a-valid-timestamp")).To(Succeed())
						})

						it("returns an error", func() {
							buffer := bytes.NewBuffer(nil)
							_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.SPDXFormat))
							Expect(err).To(MatchError(ContainSubstring("failed to parse SOURCE_DATE_EPOCH")))
						})
					})
				})
			})
		})
	}, spec.Sequential())

	it("writes the SBOM in the latest syft format", func() {
		buffer := bytes.NewBuffer(nil)
		_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.SyftFormat))
		Expect(err).NotTo(HaveOccurred())

		var syftOutput syftOutput

		err = json.Unmarshal(buffer.Bytes(), &syftOutput)
		Expect(err).NotTo(HaveOccurred(), buffer.String())

		Expect(syftOutput.Source.Type).To(Equal("directory"), buffer.String())
		Expect(syftOutput.Source.Metadata.Path).To(Equal("testdata/"), buffer.String())
		Expect(syftOutput.Artifacts[0].Name).To(Equal("collapse-white-space"), buffer.String())
		Expect(syftOutput.Artifacts[1].Name).To(Equal("end-of-stream"), buffer.String())
		Expect(syftOutput.Artifacts[2].Name).To(Equal("insert-css"), buffer.String())
		Expect(syftOutput.Artifacts[3].Name).To(Equal("once"), buffer.String())
		Expect(syftOutput.Artifacts[4].Name).To(Equal("pump"), buffer.String())
		Expect(syftOutput.Artifacts[5].Name).To(Equal("wrappy"), buffer.String())

		rerunBuffer := bytes.NewBuffer(nil)
		_, err = io.Copy(rerunBuffer, sbom.NewFormattedReader(bom, sbom.SyftFormat))
		Expect(err).NotTo(HaveOccurred())
		Expect(rerunBuffer.String()).To(Equal(buffer.String()))
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
