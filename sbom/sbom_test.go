package sbom_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSBOM(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("GenerateFromDependency", func() {
		it("generates a SBOM from a dependency", func() {
			bom, err := sbom.GenerateFromDependency(postal.Dependency{
				CPE:          "cpe:2.3:a:golang:go:1.16.9:*:*:*:*:*:*:*",
				ID:           "go",
				Licenses:     []string{"BSD-3-Clause"},
				Name:         "Go",
				PURL:         "pkg:generic/go@go1.16.9?checksum=0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d&download_url=https://dl.google.com/go/go1.16.9.src.tar.gz",
				SHA256:       "ca9ef23a5db944b116102b87c1ae9344b27e011dae7157d2f1e501abd39e9829",
				Source:       "https://dl.google.com/go/go1.16.9.src.tar.gz",
				SourceSHA256: "0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d",
				Stacks:       []string{"io.buildpacks.stacks.bionic", "io.paketo.stacks.tiny"},
				URI:          "https://deps.paketo.io/go/go_go1.16.9_linux_x64_bionic_ca9ef23a.tgz",
				Version:      "1.16.9",
			}, "some-path")
			Expect(err).NotTo(HaveOccurred())

			formatter, err := bom.InFormats(sbom.SyftFormat, sbom.CycloneDXFormat, sbom.SPDXFormat)
			Expect(err).NotTo(HaveOccurred())

			formats := formatter.Formats()

			syft := bytes.NewBuffer(nil)
			for _, format := range formats {
				if format.Extension == "syft.json" {
					_, err = io.Copy(syft, format.Content)
					Expect(err).NotTo(HaveOccurred())
				}
			}

			var syftOutput syftOutput

			err = json.Unmarshal(syft.Bytes(), &syftOutput)
			Expect(err).NotTo(HaveOccurred(), syft.String())

			goArtifact := syftOutput.Artifacts[0]
			Expect(goArtifact.Name).To(Equal("Go"), syft.String())
			Expect(goArtifact.Version).To(Equal("1.16.9"), syft.String())
			Expect(goArtifact.Licenses).To(Equal([]string{"BSD-3-Clause"}), syft.String())
			Expect(goArtifact.CPEs).To(Equal([]string{"cpe:2.3:a:golang:go:1.16.9:*:*:*:*:*:*:*"}), syft.String())
			Expect(goArtifact.PURL).To(Equal("pkg:generic/go@go1.16.9?checksum=0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d&download_url=https://dl.google.com/go/go1.16.9.src.tar.gz"), syft.String())
			Expect(syftOutput.Source.Type).To(Equal("directory"), syft.String())
			Expect(syftOutput.Source.Target).To(Equal("some-path"), syft.String())
			Expect(syftOutput.Schema.Version).To(MatchRegexp(`2\.0\.\d+`), syft.String())

			cdx := bytes.NewBuffer(nil)
			for _, format := range formats {
				if format.Extension == "cdx.json" {
					_, err = io.Copy(cdx, format.Content)
					Expect(err).NotTo(HaveOccurred())
				}
			}

			type license struct {
				License struct {
					Name string `json:"name"`
				} `json:"license"`
			}

			type component struct {
				Type     string    `json:"type"`
				Name     string    `json:"name"`
				Version  string    `json:"version"`
				Licenses []license `json:"licenses"`
				PURL     string    `json:"purl"`
			}

			var cdxOutput cdxOutput

			err = json.Unmarshal(cdx.Bytes(), &cdxOutput)
			Expect(err).NotTo(HaveOccurred(), cdx.String())

			Expect(cdxOutput.BOMFormat).To(Equal("CycloneDX"))
			Expect(cdxOutput.SpecVersion).To(Equal("1.3"))

			goComponent := cdxOutput.Components[0]
			Expect(goComponent.Name).To(Equal("Go"), cdx.String())
			Expect(goComponent.Version).To(Equal("1.16.9"), cdx.String())
			Expect(goComponent.Licenses).To(HaveLen(1), cdx.String())
			Expect(goComponent.Licenses[0].License.Name).To(Equal("BSD-3-Clause"), cdx.String())
			Expect(goComponent.PURL).To(Equal("pkg:generic/go@go1.16.9?checksum=0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d&download_url=https://dl.google.com/go/go1.16.9.src.tar.gz"), cdx.String())

			Expect(cdxOutput.Metadata.Component.Type).To(Equal("file"), cdx.String())
			Expect(cdxOutput.Metadata.Component.Name).To(Equal("some-path"), cdx.String())

			spdx := bytes.NewBuffer(nil)
			for _, format := range formats {
				if format.Extension == "spdx.json" {
					_, err = io.Copy(spdx, format.Content)
					Expect(err).NotTo(HaveOccurred())
				}
			}

			var spdxOutput spdxOutput

			err = json.Unmarshal(spdx.Bytes(), &spdxOutput)
			Expect(err).NotTo(HaveOccurred())
			Expect(err).NotTo(HaveOccurred(), spdx.String())

			Expect(spdxOutput.SPDXVersion).To(Equal("SPDX-2.2"), spdx.String())

			goPackage := spdxOutput.Packages[0]
			Expect(goPackage.Name).To(Equal("Go"), spdx.String())
			Expect(goPackage.Version).To(Equal("1.16.9"), spdx.String())
			Expect(goPackage.LicenseConcluded).To(Equal("BSD-3-Clause"), spdx.String())
			Expect(goPackage.LicenseDeclared).To(Equal("BSD-3-Clause"), spdx.String())
			Expect(goPackage.ExternalRefs).To(ContainElement(externalRef{
				Category: "SECURITY",
				Locator:  "cpe:2.3:a:golang:go:1.16.9:*:*:*:*:*:*:*",
				Type:     "cpe23Type",
			}), spdx.String())
			Expect(goPackage.ExternalRefs).To(ContainElement(externalRef{
				Category: "PACKAGE_MANAGER",
				Locator:  "pkg:generic/go@go1.16.9?checksum=0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d&download_url=https://dl.google.com/go/go1.16.9.src.tar.gz",
				Type:     "purl",
			}), spdx.String())
		})

		context("failure cases", func() {
			context("when the CPE is invalid", func() {
				it("returns an error", func() {
					_, err := sbom.GenerateFromDependency(postal.Dependency{
						CPE: "not a valid CPE",
					}, "some-path")
					Expect(err).To(MatchError(ContainSubstring("failed to parse CPE")))
				})
			})
		})
	})

	context("InFormats", func() {
		context("failure cases", func() {
			context("when a format is not supported", func() {
				it("returns an error", func() {
					_, err := sbom.SBOM{}.InFormats("unknown-format")
					Expect(err).To(MatchError(`"unknown-format" is not a supported SBOM format`))
				})
			})
		})
	})
}
