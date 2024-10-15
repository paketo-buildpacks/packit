package sbomgen_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/sbomgen"
	"github.com/paketo-buildpacks/packit/v2/sbomgen/fakes"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSyftCLIScanner(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("NewSBOMCLIScanner", func() {
		var (
			syftCLIScanner sbomgen.SyftCLIScanner
			logsBuffer     *bytes.Buffer
			layersDir      string
			err            error

			executions []pexec.Execution
			executable *fakes.Executable
		)

		it.Before(func() {
			logsBuffer = bytes.NewBuffer(nil)
			executable = &fakes.Executable{}

			layersDir, err = os.MkdirTemp("", "layers")
			Expect(err).NotTo(HaveOccurred())

			executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
				executions = append(executions, execution)
				if strings.Contains(strings.Join(execution.Args, " "), "cyclonedx-json") {
					Expect(os.WriteFile(filepath.Join(layersDir, "some-layer-name.sbom.cdx.json"), []byte(`{
  "bomFormat": "CycloneDX",
  "specVersion": "1.4",
  "serialNumber": "urn:uuid:5d2fcb74-b20f-4091-b3ce-b29201f136eb",
  "version": 1,
  "metadata": {
    "timestamp": "2024-09-09T17:28:12Z",
    "tools": [
      {
        "vendor": "anchore",
        "name": "syft",
        "version": "1.11.1"
      }
    ],
    "component": {
      "bom-ref": "5b6e90752b6334f9",
      "type": "file",
      "name": "/layers/paketo-buildpacks_node-engine/node"
    }
  }
}`), 0600)).To(Succeed())
				}

				if strings.Contains(strings.Join(execution.Args, " "), "spdx-json") {
					Expect(os.WriteFile(filepath.Join(layersDir, "some-layer-name.sbom.spdx.json"), []byte(`{
  "spdxVersion": "SPDX-2.3",
  "name": "/workspace",
  "documentNamespace": "https://anchore.com/syft/dir/workspace-2188c148-ec69-4e9c-a6c5-e24f2d738ba2",
  "creationInfo": {
    "licenseListVersion": "3.23",
    "created": "2024-08-07T17:28:12Z"
  },
  "packages": [
    {
      "name": "apackage",
      "SPDXID": "SPDXRef-Package-npm-apackage-4bc84cbb6d76f2fa",
      "versionInfo": "9.8.7",
      "downloadLocation": "https://registry.npmjs.org/apackage/-/apackage-9.8.7.tgz"
    }
  ],
  "files": [
    {
      "fileName": "/package-lock.json",
      "SPDXID": "SPDXRef-File-package-lock.json-fd71c2238fc07657"
    }
  ],
  "relationships": [
    {
      "relationshipType": "OTHER",
      "comment": "evident-by: indicates the package's existence is evident by the given file"
    }
  ]
}`), 0600)).To(Succeed())
				}
				return nil
			}

			syftCLIScanner = sbomgen.NewSyftCLIScanner(
				executable,
				scribe.NewEmitter(logsBuffer),
			)
		})

		it.After(func() {
			Expect(os.RemoveAll(layersDir)).To(Succeed())
		})

		context("GenerateSBOM", func() {
			context("syft CLI execution", func() {
				context("single mediatype without a version", func() {
					it("runs the cli commands to scan and generate SBOM", func() {
						err := syftCLIScanner.GenerateSBOM("some-path", layersDir, "some-layer-name", sbomgen.CycloneDXFormat)
						Expect(err).NotTo(HaveOccurred())

						Expect(executions).To(HaveLen(1))
						Expect(executions[0].Args).To(Equal([]string{
							"scan",
							"--quiet",
							"--output", fmt.Sprintf("cyclonedx-json=%s/some-layer-name.sbom.cdx.json", layersDir),
							"some-path",
						}))
					})
				})

				context("multiple mediatypes without a version", func() {
					it("runs the cli commands to scan and generate SBOM", func() {
						err := syftCLIScanner.GenerateSBOM("some-path", layersDir, "some-layer-name",
							sbomgen.CycloneDXFormat, sbomgen.SPDXFormat, sbomgen.SyftFormat)
						Expect(err).NotTo(HaveOccurred())

						Expect(executions).To(HaveLen(1))
						Expect(executions[0].Args).To(Equal([]string{
							"scan",
							"--quiet",
							"--output", fmt.Sprintf("cyclonedx-json=%s/some-layer-name.sbom.cdx.json", layersDir),
							"--output", fmt.Sprintf("spdx-json=%s/some-layer-name.sbom.spdx.json", layersDir),
							"--output", fmt.Sprintf("syft-json=%s/some-layer-name.sbom.syft.json", layersDir),
							"some-path",
						}))
					})
				})

				context("multiple mediatypes with and without version", func() {
					it("runs the cli commands to scan and generate SBOM", func() {
						err := syftCLIScanner.GenerateSBOM("some-path", layersDir, "some-layer-name",
							sbomgen.CycloneDXFormat+";version=1.2.3", sbomgen.SPDXFormat, sbomgen.SyftFormat)
						Expect(err).NotTo(HaveOccurred())

						Expect(executions).To(HaveLen(1))
						Expect(executions[0].Args).To(Equal([]string{
							"scan",
							"--quiet",
							"--output", fmt.Sprintf("cyclonedx-json@1.2.3=%s/some-layer-name.sbom.cdx.json", layersDir),
							"--output", fmt.Sprintf("spdx-json=%s/some-layer-name.sbom.spdx.json", layersDir),
							"--output", fmt.Sprintf("syft-json=%s/some-layer-name.sbom.syft.json", layersDir),
							"some-path",
						}))
					})
				})
			})

			context("making CLI CycloneDX output reproducible", func() {
				it("removes non-reproducible fields from CycloneDX SBOM", func() {
					err := syftCLIScanner.GenerateSBOM("some-path", layersDir, "some-layer-name", sbomgen.CycloneDXFormat)
					Expect(err).NotTo(HaveOccurred())

					generatedSBOM, err := os.ReadFile(filepath.Join(layersDir, "some-layer-name.sbom.cdx.json"))
					Expect(err).NotTo(HaveOccurred())

					// This is the stub-generated SBOM with non-repro fields removed
					expectedSBOM := `{
	"bomFormat": "CycloneDX",
  "specVersion": "1.4",
  "version": 1,
  "metadata": {
    "tools": [
      {
        "vendor": "anchore",
        "name": "syft",
        "version": "1.11.1"
      }
    ],
    "component": {
      "bom-ref": "5b6e90752b6334f9",
      "type": "file",
      "name": "/layers/paketo-buildpacks_node-engine/node"
    }
  }
}`
					var obj1, obj2 interface{}
					err = json.Unmarshal([]byte(generatedSBOM), &obj1)
					Expect(err).NotTo(HaveOccurred())
					err = json.Unmarshal([]byte(expectedSBOM), &obj2)
					Expect(err).NotTo(HaveOccurred())
					Expect(reflect.DeepEqual(obj1, obj2)).To(BeTrue())
				})
			})

			context("making CLI SPDX output reproducible", func() {
				context("without setting $SOURCE_DATE_EPOCH", func() {
					it("modifies non-reproducible fields from SPDX SBOM", func() {
						err := syftCLIScanner.GenerateSBOM("some-path", layersDir, "some-layer-name", sbomgen.SPDXFormat)
						Expect(err).NotTo(HaveOccurred())

						generatedSBOM, err := os.ReadFile(filepath.Join(layersDir, "some-layer-name.sbom.spdx.json"))
						Expect(err).NotTo(HaveOccurred())

						var generatedSBOMObj spdxOutput
						err = json.Unmarshal([]byte(generatedSBOM), &generatedSBOMObj)
						Expect(err).NotTo(HaveOccurred())

						// Ensure documentNamespace and creationInfo.created have reproducible values
						Expect(generatedSBOMObj.DocumentNamespace).To(Equal("https://paketo.io/packit/dir/workspace-b45eebde-57b8-5069-8df8-bcf8bc91810f"))
						Expect(generatedSBOMObj.CreationInfo.Created).To(BeZero())

						// Make sure others are unchanged
						Expect(generatedSBOMObj.SPDXVersion).To(Equal("SPDX-2.3"))
						Expect(generatedSBOMObj.Packages).To(HaveLen(1))
						Expect(generatedSBOMObj.Packages[0].Name).To(Equal("apackage"))
						Expect(generatedSBOMObj.Packages[0].Version).To(Equal("9.8.7"))
					})
				})

				context("setting $SOURCE_DATE_EPOCH", func() {
					var original string

					it.Before(func() {
						original = os.Getenv("SOURCE_DATE_EPOCH")
						Expect(os.Setenv("SOURCE_DATE_EPOCH", "1659551872")).To(Succeed())
					})

					it.After(func() {
						Expect(os.Setenv("SOURCE_DATE_EPOCH", original)).To(Succeed())
					})

					it("modifies non-reproducible fields from SPDX SBOM", func() {
						err := syftCLIScanner.GenerateSBOM("some-path", layersDir, "some-layer-name", sbomgen.SPDXFormat)
						Expect(err).NotTo(HaveOccurred())

						generatedSBOM, err := os.ReadFile(filepath.Join(layersDir, "some-layer-name.sbom.spdx.json"))
						Expect(err).NotTo(HaveOccurred())

						var generatedSBOMObj spdxOutput
						err = json.Unmarshal([]byte(generatedSBOM), &generatedSBOMObj)
						Expect(err).NotTo(HaveOccurred())

						// Ensure documentNamespace and creationInfo.created have reproducible values
						Expect(generatedSBOMObj.DocumentNamespace).To(Equal("https://paketo.io/packit/dir/workspace-28ef3e20-b1ec-522e-9bd5-0fcf2b7ea5c2"))
						Expect(generatedSBOMObj.CreationInfo.Created).To(Equal(time.Unix(1659551872, 0).UTC()))
					})
				})
			})

			context("failure cases", func() {
				context("invalid mediatype name", func() {
					it("shows an invalid type error", func() {
						err := syftCLIScanner.GenerateSBOM("some-path", layersDir, "some-layer-name", "whatever-mediatype")
						Expect(err).To(MatchError(ContainSubstring("mediatype whatever-mediatype matched none of the known mediatypes. Valid values are [application/vnd.cyclonedx+json application/spdx+json application/vnd.syft+json], with an optional version param for CycloneDX and SPDX")))
					})
				})

				context("invalid mediatype version format", func() {
					it("shows an invalid mediatype version format error", func() {
						err := syftCLIScanner.GenerateSBOM("some-path", layersDir, "some-layer-name", "application/vnd.cyclonedx+json;;foo")
						Expect(err).To(MatchError(ContainSubstring("Expected <mediatype>[;version=<ver>], Got application/vnd.cyclonedx+json;;foo")))
					})
				})

				context("syft mediatype contains a version specifier", func() {
					it("shows an error", func() {
						err := syftCLIScanner.GenerateSBOM("some-path", layersDir, "some-layer-name",
							sbomgen.CycloneDXFormat, sbomgen.SPDXFormat, sbomgen.SyftFormat+";version=1.2.3")
						Expect(err).To(MatchError(ContainSubstring("The syft mediatype does not allow providing a ;version=<ver> param")))
					})
				})

				context("syft CLI execution fails", func() {
					it.Before(func() {
						executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
							fmt.Fprintln(execution.Stdout, "cli error stdout")
							fmt.Fprintln(execution.Stderr, "cli error stderr")
							return fmt.Errorf("cli command failed")
						}
					})
					it("returns an error & writes to logs", func() {
						err := syftCLIScanner.GenerateSBOM("some-path", layersDir, "some-layer-name", sbomgen.CycloneDXFormat+";version=1.2.3", sbomgen.SPDXFormat, sbomgen.SyftFormat)
						Expect(err).To(MatchError(ContainSubstring(
							fmt.Sprintf("failed to execute syft cli with args '[scan --quiet --output cyclonedx-json@1.2.3=%s/some-layer-name.sbom.cdx.json --output spdx-json=%s/some-layer-name.sbom.spdx.json --output syft-json=%s/some-layer-name.sbom.syft.json some-path]'", layersDir, layersDir, layersDir))))
						Expect(err).To(MatchError(ContainSubstring("cli command failed")))
						Expect(err).To(MatchError(ContainSubstring("You might be missing a buildpack that provides the syft CLI")))

						Expect(logsBuffer.String()).To(ContainSubstring("cli error stdout"))
						Expect(logsBuffer.String()).To(ContainSubstring("cli error stderr"))
					})
				})

				context("making CycloneDX output reproducible fails", func() {
					var tmpLayersDir string
					var err error

					it.Before(func() {
						tmpLayersDir, err = os.MkdirTemp("", "layers")
						Expect(err).NotTo(HaveOccurred())
						Expect(os.WriteFile(filepath.Join(tmpLayersDir, "some-layer-name.sbom.cdx.json"), []byte(`invalid-sbom`), 0600)).To(Succeed())
					})

					it.After(func() {
						Expect(os.RemoveAll(tmpLayersDir)).To(Succeed())
					})

					it("returns helpful error message", func() {
						err := syftCLIScanner.GenerateSBOM("some-path", tmpLayersDir, "some-layer-name", sbomgen.CycloneDXFormat)
						Expect(err).To(MatchError(ContainSubstring("failed to make CycloneDX SBOM reproducible: unable to decode CycloneDX JSON")))
					})
				})

				context("making SPDX output reproducible fails", func() {
					var tmpLayersDir string
					var err error

					it.Before(func() {
						tmpLayersDir, err = os.MkdirTemp("", "layers")
						Expect(err).NotTo(HaveOccurred())
						Expect(os.WriteFile(filepath.Join(tmpLayersDir, "some-layer-name.sbom.spdx.json"), []byte(`invalid-sbom`), 0600)).To(Succeed())
					})

					it.After(func() {
						Expect(os.RemoveAll(tmpLayersDir)).To(Succeed())
					})

					it("returns helpful error message", func() {
						err := syftCLIScanner.GenerateSBOM("some-path", tmpLayersDir, "some-layer-name", sbomgen.SPDXFormat)
						Expect(err).To(MatchError(ContainSubstring("failed to make SPDX SBOM reproducible: unable to decode SPDX JSON")))
					})
				})
			})
		})
	})
}
