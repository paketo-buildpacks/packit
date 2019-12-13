package cargo_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cloudfoundry/packit/cargo"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/packit/matchers"
	. "github.com/onsi/gomega"
)

func testConfig(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer *bytes.Buffer
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
	})

	context("EncodeConfig", func() {
		it("encodes the config to TOML", func() {
			err := cargo.EncodeConfig(buffer, cargo.Config{
				API: "0.2",
				Buildpack: cargo.ConfigBuildpack{
					ID:      "some-buildpack-id",
					Name:    "some-buildpack-name",
					Version: "some-buildpack-version",
				},
				Stacks: []cargo.ConfigStack{
					{
						ID: "some-stack-id",
					},
				},
				Metadata: cargo.ConfigMetadata{
					IncludeFiles: []string{
						"some-include-file",
						"other-include-file",
					},
					Unstructured: map[string]interface{}{"some-map": []map[string]interface{}{{"key": "value"}}},
					PrePackage:   "some-pre-package-script.sh",
					Dependencies: []cargo.ConfigMetadataDependency{
						{
							ID:      "some-dependency",
							Name:    "Some Dependency",
							SHA256:  "shasum",
							Stacks:  []string{"io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"},
							URI:     "http://some-url",
							Version: "1.2.3",
						},
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(buffer.String()).To(MatchTOML(`api = "0.2"
[buildpack]
id = "some-buildpack-id"
name = "some-buildpack-name"
version = "some-buildpack-version"

[metadata]
include_files = ["some-include-file", "other-include-file"]
pre_package = "some-pre-package-script.sh"

[[metadata.dependencies]]
  id = "some-dependency"
  name = "Some Dependency"
  sha256 = "shasum"
  stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
  uri = "http://some-url"
  version = "1.2.3"

[[metadata.some-map]]
  key = "value"

[[stacks]]
  id = "some-stack-id"

`))
		})

		context("failure cases", func() {
			context("when the Config cannot be marshalled to json", func() {
				it("returns an error", func() {
					err := cargo.EncodeConfig(bytes.NewBuffer(nil), cargo.Config{
						Metadata: cargo.ConfigMetadata{
							Unstructured: map[string]interface{}{
								"some-key": func() {},
							},
						},
					})

					Expect(err).To(MatchError(ContainSubstring("json: unsupported type")))
				})
			})
		})
	})

	context("DecodeConfig", func() {
		it("decodes TOML to config", func() {
			tomlBuffer := strings.NewReader(`api = "0.2"
[buildpack]
id = "some-buildpack-id"
name = "some-buildpack-name"
version = "some-buildpack-version"

[metadata]
include_files = ["some-include-file", "other-include-file"]
pre_package = "some-pre-package-script.sh"

[[metadata.some-map]]
  key = "value"

[[metadata.dependencies]]
  id = "some-dependency"
  name = "Some Dependency"
  sha256 = "shasum"
  stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
  uri = "http://some-url"
  version = "1.2.3"

[[stacks]]
  id = "some-stack-id"
`)

			var config cargo.Config
			Expect(cargo.DecodeConfig(tomlBuffer, &config)).To(Succeed())
			Expect(config).To(Equal(cargo.Config{
				API: "0.2",
				Buildpack: cargo.ConfigBuildpack{
					ID:      "some-buildpack-id",
					Name:    "some-buildpack-name",
					Version: "some-buildpack-version",
				},
				Stacks: []cargo.ConfigStack{
					{
						ID: "some-stack-id",
					},
				},
				Metadata: cargo.ConfigMetadata{
					Unstructured: map[string]interface{}{"some-map": json.RawMessage(`[{"key":"value"}]`)},
					IncludeFiles: []string{
						"some-include-file",
						"other-include-file",
					},
					PrePackage: "some-pre-package-script.sh",
					Dependencies: []cargo.ConfigMetadataDependency{
						{
							ID:      "some-dependency",
							Name:    "Some Dependency",
							SHA256:  "shasum",
							Stacks:  []string{"io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"},
							URI:     "http://some-url",
							Version: "1.2.3",
						},
					},
				},
			}))
		})

		context("failure cases", func() {
			context("when a bad reader is passed in", func() {
				it("returns an error", func() {
					err := cargo.DecodeConfig(errorReader{}, &cargo.Config{})
					Expect(err).To(MatchError(ContainSubstring("failed to read")))
				})
			})
		})
	})

	context("ConfigMetadata", func() {
		context("UnmarshalJSON", func() {
			context("failure cases", func() {
				context("metadata field is not a object", func() {
					it("it returns an error", func() {
						var metadata cargo.ConfigMetadata
						err := metadata.UnmarshalJSON([]byte(`"some-string"`))
						Expect(err).To(MatchError(ContainSubstring("json: cannot unmarshal")))
					})
				})

				context("metadata field include_files is not a []string", func() {
					it("it returns an error", func() {
						var metadata cargo.ConfigMetadata
						err := metadata.UnmarshalJSON([]byte(`{"include_files": "some-string"}`))
						Expect(err).To(MatchError(ContainSubstring("json: cannot unmarshal")))
					})
				})

				context("metadata field pre_package is not a string", func() {
					it("it returns an error", func() {
						var metadata cargo.ConfigMetadata
						err := metadata.UnmarshalJSON([]byte(`{"pre_package": true}`))
						Expect(err).To(MatchError(ContainSubstring("json: cannot unmarshal")))
					})
				})

				context("metadata field dependencies is not an array of objects", func() {
					it("it returns an error", func() {
						var metadata cargo.ConfigMetadata
						err := metadata.UnmarshalJSON([]byte(`{"dependencies": true}`))
						Expect(err).To(MatchError(ContainSubstring("json: cannot unmarshal")))
					})
				})
			})
		})
	})
}
