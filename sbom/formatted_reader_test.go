package sbom_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/sbom"
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

	it("writes the SBOM in CycloneDX format", func() {
		buffer := bytes.NewBuffer(nil)
		_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.CycloneDXFormat))
		Expect(err).NotTo(HaveOccurred())

		type component struct {
			Name string `json:"name"`
		}

		var cdxOutput struct {
			BOMFormat   string `json:"bomFormat"`
			SpecVersion string `json:"specVersion"`
			Metadata    struct {
				Component struct {
					Type string `json:"type"`
					Name string `json:"name"`
				} `json:"component"`
			} `json:"metadata"`
			Components []component `json:"components"`
		}

		err = json.Unmarshal(buffer.Bytes(), &cdxOutput)
		Expect(err).NotTo(HaveOccurred(), buffer.String())

		Expect(cdxOutput.BOMFormat).To(Equal("CycloneDX"), buffer.String())
		Expect(cdxOutput.SpecVersion).To(Equal("1.3"), buffer.String())
		Expect(cdxOutput.Metadata.Component.Type).To(Equal("file"), buffer.String())
		Expect(cdxOutput.Metadata.Component.Name).To(Equal("testdata/"), buffer.String())
		Expect(cdxOutput.Components).To(ContainElements([]component{
			{Name: "collapse-white-space"},
			{Name: "end-of-stream"},
			{Name: "insert-css"},
			{Name: "once"},
			{Name: "pump"},
			{Name: "wrappy"},
		}), buffer.String())
	})

	it("writes the SBOM in SPDX format", func() {
		buffer := bytes.NewBuffer(nil)
		_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.SPDXFormat))
		Expect(err).NotTo(HaveOccurred())

		type pkg struct {
			Name string `json:"name"`
		}

		var spdxOutput struct {
			Packages    []pkg  `json:"packages"`
			SPDXVersion string `json:"spdxVersion"`
		}

		err = json.Unmarshal(buffer.Bytes(), &spdxOutput)
		Expect(err).NotTo(HaveOccurred(), buffer.String())

		Expect(spdxOutput.SPDXVersion).To(Equal("SPDX-2.2"), buffer.String())
		Expect(spdxOutput.Packages).To(ContainElements([]pkg{
			{Name: "collapse-white-space"},
			{Name: "end-of-stream"},
			{Name: "insert-css"},
			{Name: "once"},
			{Name: "pump"},
			{Name: "wrappy"},
		}), buffer.String())
	})

	it("writes the SBOM in Syft format", func() {
		buffer := bytes.NewBuffer(nil)
		_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.SyftFormat))
		Expect(err).NotTo(HaveOccurred())

		type artifact struct {
			Name string `json:"name"`
		}

		var syftOutput struct {
			Artifacts []artifact `json:"artifacts"`
			Source    struct {
				Type   string `json:"type"`
				Target string `json:"target"`
			} `json:"source"`
			Schema struct {
				Version string `json:"version"`
			} `json:"schema"`
		}

		err = json.Unmarshal(buffer.Bytes(), &syftOutput)
		Expect(err).NotTo(HaveOccurred(), buffer.String())

		Expect(syftOutput.Schema.Version).To(MatchRegexp(`2\.0\.\d+`), buffer.String())
		Expect(syftOutput.Source.Type).To(Equal("directory"), buffer.String())
		Expect(syftOutput.Source.Target).To(Equal("testdata/"), buffer.String())
		Expect(syftOutput.Artifacts).To(ContainElements([]artifact{
			{Name: "collapse-white-space"},
			{Name: "end-of-stream"},
			{Name: "insert-css"},
			{Name: "once"},
			{Name: "pump"},
			{Name: "wrappy"},
		}), buffer.String())
	})

	context("Read", func() {
		context("failure cases", func() {
			context("when the SBOM cannot be encoded to the given format", func() {
				it("returns an error", func() {
					formatter := sbom.NewFormattedReader(sbom.SBOM{}, sbom.Format("unknown-format"))
					_, err := formatter.Read(make([]byte, 10))
					Expect(err).To(MatchError("failed to format sbom: unsupported format: UnknownFormatOption"))
				})
			})
		})
	})
}
