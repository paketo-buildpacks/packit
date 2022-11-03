package cargo_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/packit/v2/matchers"
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
			deprecationDate, err := time.Parse(time.RFC3339, "2020-06-01T00:00:00Z")
			Expect(err).NotTo(HaveOccurred())

			err = cargo.EncodeConfig(buffer, cargo.Config{
				API: "0.6",
				Buildpack: cargo.ConfigBuildpack{
					ID:          "some-buildpack-id",
					Name:        "some-buildpack-name",
					Version:     "some-buildpack-version",
					Homepage:    "some-buildpack-homepage",
					ClearEnv:    true,
					Description: "some-buildpack-description",
					Keywords:    []string{"some-buildpack-keyword"},
					Licenses: []cargo.ConfigBuildpackLicense{
						{
							Type: "some-license-type",
							URI:  "some-license-uri",
						},
					},
					SBOMFormats: []string{"some-sbom-format"},
				},
				Stacks: []cargo.ConfigStack{
					{
						ID:     "some-stack-id",
						Mixins: []string{"some-mixin-id"},
					},
					{
						ID: "other-stack-id",
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
							Checksum:        "sha256:some-sum",
							CPE:             "some-cpe",
							PURL:            "some-purl",
							DeprecationDate: &deprecationDate,
							ID:              "some-dependency",
							Licenses:        []interface{}{"fancy-license", "fancy-license-2"},
							Name:            "Some Dependency",
							SHA256:          "shasum",
							Source:          "source",
							SourceChecksum:  "sha256:source-shasum",
							SourceSHA256:    "source-shasum",
							Stacks:          []string{"io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"},
							StripComponents: 1,
							URI:             "http://some-url",
							Version:         "1.2.3",
						},
					},
					DependencyConstraints: []cargo.ConfigMetadataDependencyConstraint{
						{
							ID:         "some-dependency",
							Constraint: "1.*",
							Patches:    1,
						},
					},
					DefaultVersions: map[string]string{
						"some-dependency": "1.2.x",
					},
				},
				Order: []cargo.ConfigOrder{
					{
						Group: []cargo.ConfigOrderGroup{
							{
								ID:      "some-dependency",
								Version: "some-version"},
							{
								ID:       "other-dependency",
								Version:  "other-version",
								Optional: true,
							},
						},
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(buffer.String()).To(MatchTOML(`
api = "0.6"

[buildpack]
	id = "some-buildpack-id"
	name = "some-buildpack-name"
	version = "some-buildpack-version"
	homepage = "some-buildpack-homepage"
	clear-env = true
	description = "some-buildpack-description"
	keywords = [ "some-buildpack-keyword" ]
	sbom-formats = [ "some-sbom-format" ]

[[buildpack.licenses]]
  type = "some-license-type"
	uri = "some-license-uri"

[metadata]
	include-files = ["some-include-file", "other-include-file"]
	pre-package = "some-pre-package-script.sh"

[metadata.default-versions]
	some-dependency = "1.2.x"

[[metadata.dependencies]]
	checksum = "sha256:some-sum"
  cpe = "some-cpe"
  purl = "some-purl"
  deprecation_date = "2020-06-01T00:00:00Z"
  id = "some-dependency"
	licenses = ["fancy-license", "fancy-license-2"]
  name = "Some Dependency"
  sha256 = "shasum"
	source = "source"
	source-checksum = "sha256:source-shasum"
  source_sha256 = "source-shasum"
  stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
  strip-components = 1
  uri = "http://some-url"
  version = "1.2.3"

[[metadata.dependency-constraints]]
  id = "some-dependency"
  constraint = "1.*"
	patches = 1

[[metadata.some-map]]
  key = "value"

[[stacks]]
  id = "some-stack-id"
  mixins = ["some-mixin-id"]

[[stacks]]
  id = "other-stack-id"

[[order]]
  [[order.group]]
	  id = "some-dependency"
		version = "some-version"

  [[order.group]]
		id = "other-dependency"
		version = "other-version"
		optional = true
`))
		})

		context("when the config dependency licenses are structured like ConfigBuildpackLicenses", func() {
			it("encodes the config to TOML", func() {
				deprecationDate, err := time.Parse(time.RFC3339, "2020-06-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())

				err = cargo.EncodeConfig(buffer, cargo.Config{
					API: "0.6",
					Buildpack: cargo.ConfigBuildpack{
						ID:       "some-buildpack-id",
						Name:     "some-buildpack-name",
						Version:  "some-buildpack-version",
						Homepage: "some-homepage-link",
						Licenses: []cargo.ConfigBuildpackLicense{
							{
								Type: "some-license-type",
								URI:  "some-license-uri",
							},
						},
					},
					Stacks: []cargo.ConfigStack{
						{
							ID:     "some-stack-id",
							Mixins: []string{"some-mixin-id"},
						},
						{
							ID: "other-stack-id",
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
								Checksum:        "sha256:some-sum",
								CPE:             "some-cpe",
								PURL:            "some-purl",
								DeprecationDate: &deprecationDate,
								ID:              "some-dependency",
								Licenses: []interface{}{
									cargo.ConfigBuildpackLicense{
										Type: "fancy-license",
										URI:  "some-license-uri",
									},
									cargo.ConfigBuildpackLicense{
										Type: "fancy-license-2",
										URI:  "some-license-uri",
									},
								},
								Name:           "Some Dependency",
								SHA256:         "shasum",
								Source:         "source",
								SourceChecksum: "sha256:source-shasum",
								SourceSHA256:   "source-shasum",
								Stacks:         []string{"io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"},
								URI:            "http://some-url",
								Version:        "1.2.3",
							},
						},
						DependencyConstraints: []cargo.ConfigMetadataDependencyConstraint{
							{
								ID:         "some-dependency",
								Constraint: "1.*",
								Patches:    1,
							},
						},
						DefaultVersions: map[string]string{
							"some-dependency": "1.2.x",
						},
					},
					Order: []cargo.ConfigOrder{
						{
							Group: []cargo.ConfigOrderGroup{
								{
									ID:      "some-dependency",
									Version: "some-version"},
								{
									ID:       "other-dependency",
									Version:  "other-version",
									Optional: true,
								},
							},
						},
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(buffer.String()).To(MatchTOML(`
api = "0.6"

[buildpack]
	id = "some-buildpack-id"
	name = "some-buildpack-name"
	version = "some-buildpack-version"
	homepage = "some-homepage-link"

[[buildpack.licenses]]
  type = "some-license-type"
	uri = "some-license-uri"

[metadata]
	include-files = ["some-include-file", "other-include-file"]
	pre-package = "some-pre-package-script.sh"

[metadata.default-versions]
	some-dependency = "1.2.x"

[[metadata.dependencies]]
	checksum = "sha256:some-sum"
  cpe = "some-cpe"
  purl = "some-purl"
  deprecation_date = "2020-06-01T00:00:00Z"
  id = "some-dependency"
  name = "Some Dependency"
  sha256 = "shasum"
	source = "source"
	source-checksum = "sha256:source-shasum"
  source_sha256 = "source-shasum"
  stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
  uri = "http://some-url"
  version = "1.2.3"

  [[metadata.dependencies.licenses]]
    type = "fancy-license"
	  uri = "some-license-uri"

	[[metadata.dependencies.licenses]]
    type = "fancy-license-2"
	  uri = "some-license-uri"

[[metadata.dependency-constraints]]
  id = "some-dependency"
  constraint = "1.*"
	patches = 1

[[metadata.some-map]]
  key = "value"

[[stacks]]
  id = "some-stack-id"
  mixins = ["some-mixin-id"]

[[stacks]]
  id = "other-stack-id"

[[order]]
  [[order.group]]
	  id = "some-dependency"
		version = "some-version"

  [[order.group]]
		id = "other-dependency"
		version = "other-version"
		optional = true
`))
			})

		})

		context("when not all metadata.dependencies include strip-components", func() {
			it("correctly encodes all elements", func() {
				err := cargo.EncodeConfig(buffer, cargo.Config{
					API: "0.6",
					Buildpack: cargo.ConfigBuildpack{
						ID: "some-buildpack-id",
					},
					Metadata: cargo.ConfigMetadata{
						Dependencies: []cargo.ConfigMetadataDependency{
							{
								ID: "some-dependency",
							},
							{
								ID:              "other-dependency",
								StripComponents: 1,
							},
						},
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(buffer.String()).To(MatchTOML(`
api = "0.6"

[buildpack]
	id = "some-buildpack-id"

[[metadata.dependencies]]
  id = "some-dependency"

[[metadata.dependencies]]
  id = "other-dependency"
	strip-components = 1
`))
			})

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

			context("when the patches in dependency constraints cannot be converted to an int", func() {
				it("returns an error", func() {
					err := cargo.EncodeConfig(bytes.NewBuffer(nil), cargo.Config{
						Metadata: cargo.ConfigMetadata{
							DependencyConstraints: []cargo.ConfigMetadataDependencyConstraint{
								{
									Constraint: "some-valid-constraint",
									ID:         "some-valid-ID",
									Patches:    0,
								},
							},
						},
					})
					Expect(err).To(MatchError(ContainSubstring("failure to assert type: unexpected data in constraint patches")))
				})
			})
		})
	})

	context("DecodeConfig", func() {
		it("decodes TOML to config", func() {
			tomlBuffer := strings.NewReader(`
api = "0.6"

[buildpack]
	id = "some-buildpack-id"
	name = "some-buildpack-name"
	version = "some-buildpack-version"
	homepage = "some-buildpack-homepage"
	clear-env = true
	description = "some-buildpack-description"
	keywords = [ "some-buildpack-keyword" ]
	sbom-formats = [ "some-sbom-format" ]

[[buildpack.licenses]]
	type = "some-license-type"
	uri = "some-license-uri"

[metadata]
	include-files = ["some-include-file", "other-include-file"]
	pre-package = "some-pre-package-script.sh"

[metadata.default-versions]
	some-dependency = "1.2.x"

[[metadata.some-map]]
  key = "value"

[[metadata.dependencies]]
	checksum = "sha256:some-sum"
  cpe = "some-cpe"
  purl = "some-purl"
  id = "some-dependency"
	licenses = ["fancy-license", "fancy-license-2"]
  name = "Some Dependency"
  sha256 = "shasum"
  source = "source"
  source-checksum = "sha256:source-shasum"
  source_sha256 = "source-shasum"
  stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
	strip-components = 1
  uri = "http://some-url"
  version = "1.2.3"

[[metadata.dependency-constraints]]
  id = "some-dependency"
  constraint = "1.*"
	patches = 1

[[stacks]]
  id = "some-stack-id"
  mixins = ["some-mixin-id"]

[[stacks]]
  id = "other-stack-id"

[[order]]
  [[order.group]]
	  id = "some-dependency"
		version = "some-version"

  [[order.group]]
		id = "other-dependency"
		version = "other-version"
		optional = true
`)

			var config cargo.Config
			Expect(cargo.DecodeConfig(tomlBuffer, &config)).To(Succeed())
			Expect(config).To(Equal(cargo.Config{
				API: "0.6",
				Buildpack: cargo.ConfigBuildpack{
					ID:          "some-buildpack-id",
					Name:        "some-buildpack-name",
					Version:     "some-buildpack-version",
					Homepage:    "some-buildpack-homepage",
					ClearEnv:    true,
					Description: "some-buildpack-description",
					Keywords:    []string{"some-buildpack-keyword"},
					Licenses: []cargo.ConfigBuildpackLicense{
						{
							Type: "some-license-type",
							URI:  "some-license-uri",
						},
					},
					SBOMFormats: []string{"some-sbom-format"},
				},
				Stacks: []cargo.ConfigStack{
					{
						ID:     "some-stack-id",
						Mixins: []string{"some-mixin-id"},
					},
					{
						ID: "other-stack-id",
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
							Checksum:        "sha256:some-sum",
							CPE:             "some-cpe",
							PURL:            "some-purl",
							ID:              "some-dependency",
							Licenses:        []interface{}{"fancy-license", "fancy-license-2"},
							Name:            "Some Dependency",
							SHA256:          "shasum",
							Source:          "source",
							SourceChecksum:  "sha256:source-shasum",
							SourceSHA256:    "source-shasum",
							Stacks:          []string{"io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"},
							StripComponents: 1,
							URI:             "http://some-url",
							Version:         "1.2.3",
						},
					},
					DependencyConstraints: []cargo.ConfigMetadataDependencyConstraint{
						{
							ID:         "some-dependency",
							Constraint: "1.*",
							Patches:    1,
						},
					},
					DefaultVersions: map[string]string{
						"some-dependency": "1.2.x",
					},
				},
				Order: []cargo.ConfigOrder{
					{
						Group: []cargo.ConfigOrderGroup{
							{
								ID:       "some-dependency",
								Version:  "some-version",
								Optional: false,
							},
							{
								ID:       "other-dependency",
								Version:  "other-version",
								Optional: true,
							},
						},
					},
				},
			}))
		})

		context("dependency license are not a list of IDs", func() {
			it("decodes TOML to config", func() {
				tomlBuffer := strings.NewReader(`
api = "0.2"

[buildpack]
	id = "some-buildpack-id"
	name = "some-buildpack-name"
	version = "some-buildpack-version"
	homepage = "some-homepage-link"

[[buildpack.licenses]]
	type = "some-license-type"
	uri = "some-license-uri"

[metadata]
	include-files = ["some-include-file", "other-include-file"]
	pre-package = "some-pre-package-script.sh"

[metadata.default-versions]
	some-dependency = "1.2.x"

[[metadata.some-map]]
  key = "value"

[[metadata.dependencies]]
	checksum = "sha256:some-sum"
  cpe = "some-cpe"
  purl = "some-purl"
  id = "some-dependency"
  name = "Some Dependency"
  sha256 = "shasum"
  source = "source"
  source-checksum = "sha256:source-shasum"
  source_sha256 = "source-shasum"
  stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
  uri = "http://some-url"
  version = "1.2.3"

  [[metadata.dependencies.licenses]]
    type = "fancy-license"
	  uri = "some-license-uri"

	[[metadata.dependencies.licenses]]
    type = "fancy-license-2"
	  uri = "some-license-uri"

[[metadata.dependency-constraints]]
  id = "some-dependency"
  constraint = "1.*"
	patches = 1

[[stacks]]
  id = "some-stack-id"
  mixins = ["some-mixin-id"]

[[stacks]]
  id = "other-stack-id"

[[order]]
  [[order.group]]
	  id = "some-dependency"
		version = "some-version"

  [[order.group]]
		id = "other-dependency"
		version = "other-version"
		optional = true
`)

				var config cargo.Config
				Expect(cargo.DecodeConfig(tomlBuffer, &config)).To(Succeed())
				Expect(config).To(Equal(cargo.Config{
					API: "0.2",
					Buildpack: cargo.ConfigBuildpack{
						ID:       "some-buildpack-id",
						Name:     "some-buildpack-name",
						Version:  "some-buildpack-version",
						Homepage: "some-homepage-link",
						Licenses: []cargo.ConfigBuildpackLicense{
							{
								Type: "some-license-type",
								URI:  "some-license-uri",
							},
						},
					},
					Stacks: []cargo.ConfigStack{
						{
							ID:     "some-stack-id",
							Mixins: []string{"some-mixin-id"},
						},
						{
							ID: "other-stack-id",
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
								Checksum: "sha256:some-sum",
								CPE:      "some-cpe",
								PURL:     "some-purl",
								ID:       "some-dependency",
								Licenses: []interface{}{
									map[string]interface{}{
										"type": "fancy-license",
										"uri":  "some-license-uri",
									},
									map[string]interface{}{
										"type": "fancy-license-2",
										"uri":  "some-license-uri",
									},
								},
								Name:           "Some Dependency",
								SHA256:         "shasum",
								Source:         "source",
								SourceChecksum: "sha256:source-shasum",
								SourceSHA256:   "source-shasum",
								Stacks:         []string{"io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"},
								URI:            "http://some-url",
								Version:        "1.2.3",
							},
						},
						DependencyConstraints: []cargo.ConfigMetadataDependencyConstraint{
							{
								ID:         "some-dependency",
								Constraint: "1.*",
								Patches:    1,
							},
						},
						DefaultVersions: map[string]string{
							"some-dependency": "1.2.x",
						},
					},
					Order: []cargo.ConfigOrder{
						{
							Group: []cargo.ConfigOrderGroup{
								{
									ID:       "some-dependency",
									Version:  "some-version",
									Optional: false,
								},
								{
									ID:       "other-dependency",
									Version:  "other-version",
									Optional: true,
								},
							},
						},
					},
				}))
			})

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
		context("MarshalJSON", func() {
			context("when the all fields are empty", func() {
				it("does not marshal any fields", func() {
					var metadata cargo.ConfigMetadata
					output, err := metadata.MarshalJSON()
					Expect(err).NotTo(HaveOccurred())
					Expect(string(output)).To(MatchJSON(`{}`))
				})
			})
		})

		context("UnmarshalJSON", func() {
			context("when the all fields are empty", func() {
				it("does not unmarshal any fields", func() {
					var metadata cargo.ConfigMetadata
					err := metadata.UnmarshalJSON([]byte(`{}`))
					Expect(err).NotTo(HaveOccurred())
					Expect(metadata).To(Equal(cargo.ConfigMetadata{}))
				})
			})

			context("failure cases", func() {
				context("metadata field is not a object", func() {
					it("it returns an error", func() {
						var metadata cargo.ConfigMetadata
						err := metadata.UnmarshalJSON([]byte(`"some-string"`))
						Expect(err).To(MatchError(ContainSubstring("json: cannot unmarshal")))
					})
				})

				context("metadata field include-files is not a []string", func() {
					it("it returns an error", func() {
						var metadata cargo.ConfigMetadata
						err := metadata.UnmarshalJSON([]byte(`{"include-files": "some-string"}`))
						Expect(err).To(MatchError(ContainSubstring("json: cannot unmarshal")))
					})
				})

				context("metadata field pre-package is not a string", func() {
					it("it returns an error", func() {
						var metadata cargo.ConfigMetadata
						err := metadata.UnmarshalJSON([]byte(`{"pre-package": true}`))
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

				context("metadata field dependency-constraints is not an array of objects", func() {
					it("it returns an error", func() {
						var metadata cargo.ConfigMetadata
						err := metadata.UnmarshalJSON([]byte(`{"dependency-constraints": true}`))
						Expect(err).To(MatchError(ContainSubstring("json: cannot unmarshal")))
					})
				})
			})
		})
	})
}
