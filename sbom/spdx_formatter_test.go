package sbom_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/paketo-buildpacks/packit/sbom"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSPDXFormatter(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	it("writes the SBOM in SPDX format", func() {
		bom, err := sbom.Generate("testdata/")
		Expect(err).NotTo(HaveOccurred())

		buffer := bytes.NewBuffer(nil)
		_, err = io.Copy(buffer, bom.Format(sbom.SPDXFormat))
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
					"SPDXID": "SPDXRef-Package-npm-collapse-white-space-2.0.0",
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
					"sourceInfo": "acquired package info from installed node module manifest file: testdata/package-lock.json",
					"versionInfo": "2.0.0"
				},
				{
					"SPDXID": "SPDXRef-Package-npm-end-of-stream-1.4.4",
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
					"sourceInfo": "acquired package info from installed node module manifest file: testdata/package-lock.json",
					"versionInfo": "1.4.4"
				},
				{
					"SPDXID": "SPDXRef-Package-npm-insert-css-2.0.0",
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
					"sourceInfo": "acquired package info from installed node module manifest file: testdata/package-lock.json",
					"versionInfo": "2.0.0"
				},
				{
					"SPDXID": "SPDXRef-Package-npm-once-1.4.0",
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
					"sourceInfo": "acquired package info from installed node module manifest file: testdata/package-lock.json",
					"versionInfo": "1.4.0"
				},
				{
					"SPDXID": "SPDXRef-Package-npm-pump-3.0.0",
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
					"sourceInfo": "acquired package info from installed node module manifest file: testdata/package-lock.json",
					"versionInfo": "3.0.0"
				},
				{
					"SPDXID": "SPDXRef-Package-npm-wrappy-1.0.2",
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
					"sourceInfo": "acquired package info from installed node module manifest file: testdata/package-lock.json",
					"versionInfo": "1.0.2"
				}
			]
		}`, output.CreationInfo.Created, output.DocumentNamespace)))
	})
}
