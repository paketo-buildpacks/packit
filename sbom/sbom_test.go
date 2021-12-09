package sbom_test

import (
	"bytes"
	"encoding/json"
	"fmt"
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

			Expect(syft.String()).To(MatchJSON(`{
				"artifacts": [
					{
						"id": "b0a2cd11c0e13e43",
						"name": "Go",
						"version": "1.16.9",
						"type": "",
						"foundBy": "",
						"locations": [],
						"licenses": [
							"BSD-3-Clause"
						],
						"language": "",
						"cpes": [
							"cpe:2.3:a:golang:go:1.16.9:*:*:*:*:*:*:*"
						],
						"purl": "pkg:generic/go@go1.16.9?checksum=0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d&download_url=https://dl.google.com/go/go1.16.9.src.tar.gz",
						"metadataType": "",
						"metadata": null
					}
				],
				"artifactRelationships": [],
				"source": {
					"type": "directory",
					"target": "some-path"
				},
				"distro": {
					"name": "",
					"version": "",
					"idLike": ""
				},
				"descriptor": {
					"name": "",
					"version": ""
				},
				"schema": {
					"version": "2.0.1",
					"url": "https://raw.githubusercontent.com/anchore/syft/main/schema/json/schema-2.0.1.json"
				}
			}`))

			cdx := bytes.NewBuffer(nil)
			for _, format := range formats {
				if format.Extension == "cdx.json" {
					_, err = io.Copy(cdx, format.Content)
					Expect(err).NotTo(HaveOccurred())
				}
			}

			var cdxOutput struct {
				SerialNumber string `json:"serialNumber"`
				Metadata     struct {
					Timestamp string `json:"timestamp"`
				} `json:"metadata"`
			}
			err = json.Unmarshal(cdx.Bytes(), &cdxOutput)
			Expect(err).NotTo(HaveOccurred())

			Expect(cdx.String()).To(MatchJSON(fmt.Sprintf(`{
				"bomFormat": "CycloneDX",
				"specVersion": "1.3",
				"version": 1,
				"serialNumber": "%s",
				"metadata": {
					"timestamp": "%s",
					"tools": [
						{
							"vendor": "anchore",
							"name": "syft",
							"version": "[not provided]"
						}
					],
					"component": {
						"type": "file",
						"name": "some-path",
						"version": ""
					}
				},
				"components": [
					{
						"type": "library",
						"name": "Go",
						"version": "1.16.9",
						"licenses": [
							{
								"license": {
									"name": "BSD-3-Clause"
								}
							}
						],
						"purl": "pkg:generic/go@go1.16.9?checksum=0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d&download_url=https://dl.google.com/go/go1.16.9.src.tar.gz"
					}
				]
			}`, cdxOutput.SerialNumber, cdxOutput.Metadata.Timestamp)))

			spdx := bytes.NewBuffer(nil)
			for _, format := range formats {
				if format.Extension == "spdx.json" {
					_, err = io.Copy(spdx, format.Content)
					Expect(err).NotTo(HaveOccurred())
				}
			}

			var spdxOutput struct {
				CreationInfo struct {
					Created string `json:"created"`
				} `json:"creationInfo"`
				DocumentNamespace string `json:"documentNamespace"`
			}
			err = json.Unmarshal(spdx.Bytes(), &spdxOutput)
			Expect(err).NotTo(HaveOccurred())

			Expect(spdx.String()).To(MatchJSON(fmt.Sprintf(`{
				"SPDXID": "SPDXRef-DOCUMENT",
				"name": "some-path",
				"spdxVersion": "SPDX-2.2",
				"creationInfo": {
					"created": "%s",
					"creators": [
						"Organization: Anchore, Inc",
						"Tool: syft-[not provided]"
					],
					"licenseListVersion": "3.15"
				},
				"dataLicense": "CC0-1.0",
				"documentNamespace": "%s",
				"packages": [
					{
						"SPDXID": "SPDXRef-b0a2cd11c0e13e43",
						"name": "Go",
						"licenseConcluded": "BSD-3-Clause",
						"downloadLocation": "NOASSERTION",
						"externalRefs": [
							{
								"referenceCategory": "SECURITY",
								"referenceLocator": "cpe:2.3:a:golang:go:1.16.9:*:*:*:*:*:*:*",
								"referenceType": "cpe23Type"
							},
							{
								"referenceCategory": "PACKAGE_MANAGER",
								"referenceLocator": "pkg:generic/go@go1.16.9?checksum=0a1cc7fd7bd20448f71ebed64d846138850d5099b18cf5cc10a4fc45160d8c3d&download_url=https://dl.google.com/go/go1.16.9.src.tar.gz",
								"referenceType": "purl"
							}
						],
						"filesAnalyzed": false,
						"licenseDeclared": "BSD-3-Clause",
						"sourceInfo": "acquired package info from the following paths: ",
						"versionInfo": "1.16.9"
					}
				]
			}`, spdxOutput.CreationInfo.Created, spdxOutput.DocumentNamespace)))
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
