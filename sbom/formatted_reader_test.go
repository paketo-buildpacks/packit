package sbom_test

import (
	"bytes"
	"encoding/json"
	"fmt"
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

		var output struct {
			SerialNumber string `json:"serialNumber"`
			Metadata     struct {
				Timestamp string `json:"timestamp"`
			} `json:"metadata"`
		}
		err = json.Unmarshal(buffer.Bytes(), &output)
		Expect(err).NotTo(HaveOccurred())

		Expect(buffer.String()).To(MatchJSON(fmt.Sprintf(`{
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
					"name": "testdata/",
					"version": ""
				}
			},
			"components": [
				{
					"type": "library",
					"name": "collapse-white-space",
					"version": "2.0.0",
					"purl": "pkg:npm/collapse-white-space@2.0.0"
				},
				{
					"type": "library",
					"name": "end-of-stream",
					"version": "1.4.4",
					"purl": "pkg:npm/end-of-stream@1.4.4"
				},
				{
					"type": "library",
					"name": "insert-css",
					"version": "2.0.0",
					"purl": "pkg:npm/insert-css@2.0.0"
				},
				{
					"type": "library",
					"name": "once",
					"version": "1.4.0",
					"purl": "pkg:npm/once@1.4.0"
				},
				{
					"type": "library",
					"name": "pump",
					"version": "3.0.0",
					"purl": "pkg:npm/pump@3.0.0"
				},
				{
					"type": "library",
					"name": "wrappy",
					"version": "1.0.2",
					"purl": "pkg:npm/wrappy@1.0.2"
				}
			]
		}`, output.SerialNumber, output.Metadata.Timestamp)))
	})

	it("writes the SBOM in SPDX format", func() {
		buffer := bytes.NewBuffer(nil)
		_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.SPDXFormat))
		Expect(err).NotTo(HaveOccurred())

		var output struct {
			CreationInfo struct {
				Created string `json:"created"`
			} `json:"creationInfo"`
			DocumentNamespace string `json:"documentNamespace"`
		}
		err = json.Unmarshal(buffer.Bytes(), &output)
		Expect(err).NotTo(HaveOccurred())

		Expect(buffer.String()).To(MatchJSON(fmt.Sprintf(`{
			"SPDXID": "SPDXRef-DOCUMENT",
			"name": "testdata",
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
					"SPDXID": "SPDXRef-8cb201e31f86deb7",
					"name": "collapse-white-space",
					"licenseConcluded": "NONE",
					"downloadLocation": "NOASSERTION",
					"externalRefs": [
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:collapse-white-space:collapse-white-space:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:collapse-white-space:collapse_white_space:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:collapse_white_space:collapse-white-space:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:collapse_white_space:collapse_white_space:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:collapse-white:collapse-white-space:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:collapse-white:collapse_white_space:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:collapse_white:collapse-white-space:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:collapse_white:collapse_white_space:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:collapse:collapse-white-space:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:collapse:collapse_white_space:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:*:collapse-white-space:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:*:collapse_white_space:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "PACKAGE_MANAGER",
							"referenceLocator": "pkg:npm/collapse-white-space@2.0.0",
							"referenceType": "purl"
						}
					],
					"filesAnalyzed": false,
					"licenseDeclared": "NONE",
					"sourceInfo": "acquired package info from installed node module manifest file: package-lock.json",
					"versionInfo": "2.0.0"
				},
				{
					"SPDXID": "SPDXRef-1d5b8a47b0a05689",
					"name": "end-of-stream",
					"licenseConcluded": "NONE",
					"downloadLocation": "NOASSERTION",
					"externalRefs": [
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:end-of-stream:end-of-stream:1.4.4:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:end-of-stream:end_of_stream:1.4.4:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:end_of_stream:end-of-stream:1.4.4:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:end_of_stream:end_of_stream:1.4.4:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:end-of:end-of-stream:1.4.4:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:end-of:end_of_stream:1.4.4:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:end_of:end-of-stream:1.4.4:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:end_of:end_of_stream:1.4.4:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:end:end-of-stream:1.4.4:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:end:end_of_stream:1.4.4:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:*:end-of-stream:1.4.4:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:*:end_of_stream:1.4.4:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "PACKAGE_MANAGER",
							"referenceLocator": "pkg:npm/end-of-stream@1.4.4",
							"referenceType": "purl"
						}
					],
					"filesAnalyzed": false,
					"licenseDeclared": "NONE",
					"sourceInfo": "acquired package info from installed node module manifest file: package-lock.json",
					"versionInfo": "1.4.4"
				},
				{
					"SPDXID": "SPDXRef-fa3e7e7a7b2d55de",
					"name": "insert-css",
					"licenseConcluded": "NONE",
					"downloadLocation": "NOASSERTION",
					"externalRefs": [
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:insert-css:insert-css:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:insert-css:insert_css:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:insert_css:insert-css:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:insert_css:insert_css:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:insert:insert-css:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:insert:insert_css:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:*:insert-css:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:*:insert_css:2.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "PACKAGE_MANAGER",
							"referenceLocator": "pkg:npm/insert-css@2.0.0",
							"referenceType": "purl"
						}
					],
					"filesAnalyzed": false,
					"licenseDeclared": "NONE",
					"sourceInfo": "acquired package info from installed node module manifest file: package-lock.json",
					"versionInfo": "2.0.0"
				},
				{
					"SPDXID": "SPDXRef-76656a4e0fba5618",
					"name": "once",
					"licenseConcluded": "NONE",
					"downloadLocation": "NOASSERTION",
					"externalRefs": [
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:once:once:1.4.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:*:once:1.4.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "PACKAGE_MANAGER",
							"referenceLocator": "pkg:npm/once@1.4.0",
							"referenceType": "purl"
						}
					],
					"filesAnalyzed": false,
					"licenseDeclared": "NONE",
					"sourceInfo": "acquired package info from installed node module manifest file: package-lock.json",
					"versionInfo": "1.4.0"
				},
				{
					"SPDXID": "SPDXRef-18f9a28462e28c2e",
					"name": "pump",
					"licenseConcluded": "NONE",
					"downloadLocation": "NOASSERTION",
					"externalRefs": [
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:pump:pump:3.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:*:pump:3.0.0:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "PACKAGE_MANAGER",
							"referenceLocator": "pkg:npm/pump@3.0.0",
							"referenceType": "purl"
						}
					],
					"filesAnalyzed": false,
					"licenseDeclared": "NONE",
					"sourceInfo": "acquired package info from installed node module manifest file: package-lock.json",
					"versionInfo": "3.0.0"
				},
				{
					"SPDXID": "SPDXRef-6125f0accc5f06cf",
					"name": "wrappy",
					"licenseConcluded": "NONE",
					"downloadLocation": "NOASSERTION",
					"externalRefs": [
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:wrappy:wrappy:1.0.2:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "SECURITY",
							"referenceLocator": "cpe:2.3:a:*:wrappy:1.0.2:*:*:*:*:*:*:*",
							"referenceType": "cpe23Type"
						},
						{
							"referenceCategory": "PACKAGE_MANAGER",
							"referenceLocator": "pkg:npm/wrappy@1.0.2",
							"referenceType": "purl"
						}
					],
					"filesAnalyzed": false,
					"licenseDeclared": "NONE",
					"sourceInfo": "acquired package info from installed node module manifest file: package-lock.json",
					"versionInfo": "1.0.2"
				}
			]
		}`, output.CreationInfo.Created, output.DocumentNamespace)))
	})

	it("writes the SBOM in Syft format", func() {
		buffer := bytes.NewBuffer(nil)
		_, err := io.Copy(buffer, sbom.NewFormattedReader(bom, sbom.SyftFormat))
		Expect(err).NotTo(HaveOccurred())

		Expect(buffer.String()).To(MatchJSON(`{
			"artifacts": [
				{
					"id": "8cb201e31f86deb7",
					"name": "collapse-white-space",
					"version": "2.0.0",
					"type": "npm",
					"foundBy": "javascript-lock-cataloger",
					"locations": [
						{
							"path": "package-lock.json"
						}
					],
					"licenses": [],
					"language": "javascript",
					"cpes": [
						"cpe:2.3:a:collapse-white-space:collapse-white-space:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:collapse-white-space:collapse_white_space:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:collapse_white_space:collapse-white-space:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:collapse_white_space:collapse_white_space:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:collapse-white:collapse-white-space:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:collapse-white:collapse_white_space:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:collapse_white:collapse-white-space:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:collapse_white:collapse_white_space:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:collapse:collapse-white-space:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:collapse:collapse_white_space:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:*:collapse-white-space:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:*:collapse_white_space:2.0.0:*:*:*:*:*:*:*"
					],
					"purl": "pkg:npm/collapse-white-space@2.0.0",
					"metadataType": "",
					"metadata": null
				},
				{
					"id": "1d5b8a47b0a05689",
					"name": "end-of-stream",
					"version": "1.4.4",
					"type": "npm",
					"foundBy": "javascript-lock-cataloger",
					"locations": [
						{
							"path": "package-lock.json"
						}
					],
					"licenses": [],
					"language": "javascript",
					"cpes": [
						"cpe:2.3:a:end-of-stream:end-of-stream:1.4.4:*:*:*:*:*:*:*",
						"cpe:2.3:a:end-of-stream:end_of_stream:1.4.4:*:*:*:*:*:*:*",
						"cpe:2.3:a:end_of_stream:end-of-stream:1.4.4:*:*:*:*:*:*:*",
						"cpe:2.3:a:end_of_stream:end_of_stream:1.4.4:*:*:*:*:*:*:*",
						"cpe:2.3:a:end-of:end-of-stream:1.4.4:*:*:*:*:*:*:*",
						"cpe:2.3:a:end-of:end_of_stream:1.4.4:*:*:*:*:*:*:*",
						"cpe:2.3:a:end_of:end-of-stream:1.4.4:*:*:*:*:*:*:*",
						"cpe:2.3:a:end_of:end_of_stream:1.4.4:*:*:*:*:*:*:*",
						"cpe:2.3:a:end:end-of-stream:1.4.4:*:*:*:*:*:*:*",
						"cpe:2.3:a:end:end_of_stream:1.4.4:*:*:*:*:*:*:*",
						"cpe:2.3:a:*:end-of-stream:1.4.4:*:*:*:*:*:*:*",
						"cpe:2.3:a:*:end_of_stream:1.4.4:*:*:*:*:*:*:*"
					],
					"purl": "pkg:npm/end-of-stream@1.4.4",
					"metadataType": "",
					"metadata": null
				},
				{
					"id": "fa3e7e7a7b2d55de",
					"name": "insert-css",
					"version": "2.0.0",
					"type": "npm",
					"foundBy": "javascript-lock-cataloger",
					"locations": [
						{
							"path": "package-lock.json"
						}
					],
					"licenses": [],
					"language": "javascript",
					"cpes": [
						"cpe:2.3:a:insert-css:insert-css:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:insert-css:insert_css:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:insert_css:insert-css:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:insert_css:insert_css:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:insert:insert-css:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:insert:insert_css:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:*:insert-css:2.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:*:insert_css:2.0.0:*:*:*:*:*:*:*"
					],
					"purl": "pkg:npm/insert-css@2.0.0",
					"metadataType": "",
					"metadata": null
				},
				{
					"id": "76656a4e0fba5618",
					"name": "once",
					"version": "1.4.0",
					"type": "npm",
					"foundBy": "javascript-lock-cataloger",
					"locations": [
						{
							"path": "package-lock.json"
						}
					],
					"licenses": [],
					"language": "javascript",
					"cpes": [
						"cpe:2.3:a:once:once:1.4.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:*:once:1.4.0:*:*:*:*:*:*:*"
					],
					"purl": "pkg:npm/once@1.4.0",
					"metadataType": "",
					"metadata": null
				},
				{
					"id": "18f9a28462e28c2e",
					"name": "pump",
					"version": "3.0.0",
					"type": "npm",
					"foundBy": "javascript-lock-cataloger",
					"locations": [
						{
							"path": "package-lock.json"
						}
					],
					"licenses": [],
					"language": "javascript",
					"cpes": [
						"cpe:2.3:a:pump:pump:3.0.0:*:*:*:*:*:*:*",
						"cpe:2.3:a:*:pump:3.0.0:*:*:*:*:*:*:*"
					],
					"purl": "pkg:npm/pump@3.0.0",
					"metadataType": "",
					"metadata": null
				},
				{
					"id": "6125f0accc5f06cf",
					"name": "wrappy",
					"version": "1.0.2",
					"type": "npm",
					"foundBy": "javascript-lock-cataloger",
					"locations": [
						{
							"path": "package-lock.json"
						}
					],
					"licenses": [],
					"language": "javascript",
					"cpes": [
						"cpe:2.3:a:wrappy:wrappy:1.0.2:*:*:*:*:*:*:*",
						"cpe:2.3:a:*:wrappy:1.0.2:*:*:*:*:*:*:*"
					],
					"purl": "pkg:npm/wrappy@1.0.2",
					"metadataType": "",
					"metadata": null
				}
			],
			"artifactRelationships": [],
			"source": {
				"type": "directory",
				"target": "testdata/"
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
