package sbom_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/paketo-buildpacks/packit/sbom"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSyftFormatter(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	it("writes the SBOM in Syft format", func() {
		bom, err := sbom.Generate("testdata/")
		Expect(err).NotTo(HaveOccurred())

		buffer := bytes.NewBuffer(nil)
		_, err = io.Copy(buffer, bom.Format(sbom.SyftFormat))
		Expect(err).NotTo(HaveOccurred())

		Expect(buffer.String()).To(MatchJSON(`{
			"artifacts": [
				{
					"id": "b5116577839f53ee",
					"name": "collapse-white-space",
					"version": "2.0.0",
					"type": "npm",
					"foundBy": "javascript-lock-cataloger",
					"locations": [
						{
							"path": "testdata/package-lock.json"
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
					"id": "76de38f5dc172765",
					"name": "end-of-stream",
					"version": "1.4.4",
					"type": "npm",
					"foundBy": "javascript-lock-cataloger",
					"locations": [
						{
							"path": "testdata/package-lock.json"
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
					"id": "f5ac627dd1855cc2",
					"name": "insert-css",
					"version": "2.0.0",
					"type": "npm",
					"foundBy": "javascript-lock-cataloger",
					"locations": [
						{
							"path": "testdata/package-lock.json"
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
					"id": "4969a085850a55e4",
					"name": "once",
					"version": "1.4.0",
					"type": "npm",
					"foundBy": "javascript-lock-cataloger",
					"locations": [
						{
							"path": "testdata/package-lock.json"
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
					"id": "fce9bafbedd6e2a2",
					"name": "pump",
					"version": "3.0.0",
					"type": "npm",
					"foundBy": "javascript-lock-cataloger",
					"locations": [
						{
							"path": "testdata/package-lock.json"
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
					"id": "f1a20b83d0b6877e",
					"name": "wrappy",
					"version": "1.0.2",
					"type": "npm",
					"foundBy": "javascript-lock-cataloger",
					"locations": [
						{
							"path": "testdata/package-lock.json"
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
				"name": "syft",
				"version": "[not provided]"
			},
			"schema": {
				"version": "1.1.0",
				"url": "https://raw.githubusercontent.com/anchore/syft/main/schema/json/schema-1.1.0.json"
			}
		}`))
	})
}
