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

func testCycloneDXFormatter(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	it("writes the SBOM in CycloneDX format", func() {
		bom, err := sbom.Generate("testdata/")
		Expect(err).NotTo(HaveOccurred())

		buffer := bytes.NewBuffer(nil)
		_, err = io.Copy(buffer, bom.Format(sbom.CycloneDXFormat))
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
					"version": "",
					"licenses": null
				}
			},
			"components": [
				{
					"type": "library",
					"name": "collapse-white-space",
					"version": "2.0.0",
					"licenses": null,
					"purl": "pkg:npm/collapse-white-space@2.0.0"
				},
				{
					"type": "library",
					"name": "end-of-stream",
					"version": "1.4.4",
					"licenses": null,
					"purl": "pkg:npm/end-of-stream@1.4.4"
				},
				{
					"type": "library",
					"name": "insert-css",
					"version": "2.0.0",
					"licenses": null,
					"purl": "pkg:npm/insert-css@2.0.0"
				},
				{
					"type": "library",
					"name": "once",
					"version": "1.4.0",
					"licenses": null,
					"purl": "pkg:npm/once@1.4.0"
				},
				{
					"type": "library",
					"name": "pump",
					"version": "3.0.0",
					"licenses": null,
					"purl": "pkg:npm/pump@3.0.0"
				},
				{
					"type": "library",
					"name": "wrappy",
					"version": "1.0.2",
					"licenses": null,
					"purl": "pkg:npm/wrappy@1.0.2"
				}
			]
		}`, output.SerialNumber, output.Metadata.Timestamp)))
	})
}
