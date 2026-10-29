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
				Targets: []cargo.ConfigTarget{
					{
						OS:   "linux",
						Arch: "arm64",
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
							Checksum: "sha256:some-sum",
							CPE:      "some-cpe",
							CPEs:     []string{"some-cpe", "other-cpe"},
							PURL:     "some-purl",
							PURLs:    []string{"some-purl", "other-purl"},
							EOLDate:  &deprecationDate,
							ID:       "some-dependency",
							LicenseEntries: []cargo.ConfigDependencyLicense{
								{Type: "fancy-license"},
								{Type: "fancy-license-2"},
							},
							Name:            "Some Dependency",
							SHA256:          "shasum",
							Source:          "source",
							SourceChecksum:  "sha256:source-shasum",
							SourceSHA256:    "source-shasum",
							Stacks:          []string{"io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"},
							StripComponents: 1,
							URI:             "http://some-url",
							Version:         "1.2.3",
							OS:              "some-os",
							Arch:            "some-arch",
							Distros: []cargo.ConfigDistro{
								{
									Name:    "some-distro-name",
									Version: "some-distro-version",
								},
							},
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
  cpes = ["some-cpe", "other-cpe"]
  purl = "some-purl"
  purls = ["some-purl", "other-purl"]
  eol-date = "2020-06-01T00:00:00Z"
  id = "some-dependency"
  name = "Some Dependency"
  sha256 = "shasum"
	source = "source"
	source-checksum = "sha256:source-shasum"
  source-sha256 = "source-shasum"
  stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
  strip-components = 1
  uri = "http://some-url"
  version = "1.2.3"
  os = "some-os"
  arch = "some-arch"

	[[metadata.dependencies.distros]]
	name = "some-distro-name"
	version = "some-distro-version"

  [[metadata.dependencies.licenses]]
  type = "fancy-license"

  [[metadata.dependencies.licenses]]
  type = "fancy-license-2"

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

[[targets]]
  os = "linux"
  arch = "arm64"

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
								Checksum: "sha256:some-sum",
								CPE:      "some-cpe",
								PURL:     "some-purl",
								EOLDate:  &deprecationDate,
								ID:       "some-dependency",
								LicenseEntries: []cargo.ConfigDependencyLicense{
									{
										Type: "fancy-license",
										URI:  "some-license-uri",
									},
									{
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
  eol-date = "2020-06-01T00:00:00Z"
  id = "some-dependency"
  name = "Some Dependency"
  sha256 = "shasum"
	source = "source"
	source-checksum = "sha256:source-shasum"
  source-sha256 = "source-shasum"
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
  cpes = ["some-cpe", "other-cpe"]
  purls = ["some-purl", "other-purl"]
  eol-date = "2020-06-01T00:00:00Z"
  id = "some-dependency"
  name = "Some Dependency"
  sha256 = "shasum"
  source = "source"
  source-checksum = "sha256:source-shasum"
  source_sha256 = "source-shasum"
  stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
	strip-components = 1
  uri = "http://some-url"
  version = "1.2.3"
  os = "some-os"
  arch = "some-arch"

	[[metadata.dependencies.distros]]
	name = "some-distro-name"
	version = "some-distro-version"

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

			eolDate, err := time.Parse(time.RFC3339, "2020-06-01T00:00:00Z")
			Expect(err).NotTo(HaveOccurred())

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
							Checksum: "sha256:some-sum",
							CPEs:     []string{"some-cpe", "other-cpe"},
							PURLs:    []string{"some-purl", "other-purl"},
							EOLDate:  &eolDate,
							ID:       "some-dependency",
							LicenseEntries: []cargo.ConfigDependencyLicense{
								{Type: "fancy-license", URI: "some-license-uri"},
								{Type: "fancy-license-2", URI: "some-license-uri"},
							},
							Name:            "Some Dependency",
							SHA256:          "shasum",
							Source:          "source",
							SourceChecksum:  "sha256:source-shasum",
							SourceSHA256:    "source-shasum",
							Stacks:          []string{"io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"},
							StripComponents: 1,
							URI:             "http://some-url",
							Version:         "1.2.3",
							OS:              "some-os",
							Arch:            "some-arch",
							Distros: []cargo.ConfigDistro{
								{
									Name:    "some-distro-name",
									Version: "some-distro-version",
								},
							},
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
								LicenseEntries: []cargo.ConfigDependencyLicense{
									{
										Type: "fancy-license",
										URI:  "some-license-uri",
									},
									{
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

	// The backwards compatibility contexts below cover the dependency metadata
	// keys that predate the standard dependency metadata format (RFC 0059).
	// They exist to guarantee that buildpack.toml files written in the older
	// format keep working. Remove them once the legacy keys are no longer
	// supported.
	context("backwards compatibility", func() {
		context("DecodeConfig", func() {
			context("when the dependency uses the legacy cpe and purl keys", func() {
				it("decodes them into CPE and PURL", func() {
					tomlBuffer := strings.NewReader(`
[[metadata.dependencies]]
	id = "some-dependency"
	cpe = "some-cpe"
	purl = "some-purl"
	uri = "http://some-url"
	version = "1.2.3"
`)
					var config cargo.Config
					Expect(cargo.DecodeConfig(tomlBuffer, &config)).To(Succeed())

					Expect(config.Metadata.Dependencies).To(HaveLen(1))
					Expect(config.Metadata.Dependencies[0].CPE).To(Equal("some-cpe"))
					Expect(config.Metadata.Dependencies[0].CPEs).To(BeEmpty())
					Expect(config.Metadata.Dependencies[0].PURL).To(Equal("some-purl"))
					Expect(config.Metadata.Dependencies[0].PURLs).To(BeEmpty())
				})
			})

			context("when the dependency uses the legacy deprecation_date key", func() {
				it("decodes it into DeprecationDate", func() {
					tomlBuffer := strings.NewReader(`
[[metadata.dependencies]]
	id = "some-dependency"
	deprecation_date = "2020-06-01T00:00:00Z"
	uri = "http://some-url"
	version = "1.2.3"
`)
					eolDate, err := time.Parse(time.RFC3339, "2020-06-01T00:00:00Z")
					Expect(err).NotTo(HaveOccurred())

					var config cargo.Config
					Expect(cargo.DecodeConfig(tomlBuffer, &config)).To(Succeed())

					Expect(config.Metadata.Dependencies).To(HaveLen(1))
					Expect(config.Metadata.Dependencies[0].DeprecationDate).To(Equal(&eolDate))
				})
			})

			context("when the dependency uses the standard eol-date key", func() {
				it("decodes it into EOLDate", func() {
					tomlBuffer := strings.NewReader(`
[[metadata.dependencies]]
	id = "some-dependency"
	eol-date = "2020-06-01T00:00:00Z"
	uri = "http://some-url"
	version = "1.2.3"
`)
					eolDate, err := time.Parse(time.RFC3339, "2020-06-01T00:00:00Z")
					Expect(err).NotTo(HaveOccurred())

					var config cargo.Config
					Expect(cargo.DecodeConfig(tomlBuffer, &config)).To(Succeed())

					Expect(config.Metadata.Dependencies).To(HaveLen(1))
					Expect(config.Metadata.Dependencies[0].EOLDate).To(Equal(&eolDate))
				})
			})

			context("when the dependency uses the legacy bare string licenses form", func() {
				it("decodes them into Licenses", func() {
					tomlBuffer := strings.NewReader(`
[[metadata.dependencies]]
	id = "some-dependency"
	licenses = ["fancy-license", "fancy-license-2"]
	uri = "http://some-url"
	version = "1.2.3"
`)
					var config cargo.Config
					Expect(cargo.DecodeConfig(tomlBuffer, &config)).To(Succeed())

					Expect(config.Metadata.Dependencies).To(HaveLen(1))
					Expect(config.Metadata.Dependencies[0].Licenses).To(Equal([]interface{}{
						"fancy-license",
						"fancy-license-2",
					}))
					Expect(config.Metadata.Dependencies[0].LicenseEntries).To(BeEmpty())
				})
			})

			context("when the dependency uses the standard licenses table form", func() {
				it("decodes them into LicenseEntries", func() {
					tomlBuffer := strings.NewReader(`
[[metadata.dependencies]]
	id = "some-dependency"
	uri = "http://some-url"
	version = "1.2.3"

	[[metadata.dependencies.licenses]]
	type = "fancy-license"
	uri = "http://some-url/license"
`)
					var config cargo.Config
					Expect(cargo.DecodeConfig(tomlBuffer, &config)).To(Succeed())

					Expect(config.Metadata.Dependencies).To(HaveLen(1))
					Expect(config.Metadata.Dependencies[0].Licenses).To(BeEmpty())
					Expect(config.Metadata.Dependencies[0].LicenseEntries).To(Equal([]cargo.ConfigDependencyLicense{
						{Type: "fancy-license", URI: "http://some-url/license"},
					}))
				})
			})

			context("when the dependency uses the legacy sha256 and source_sha256 keys", func() {
				it("decodes them", func() {
					tomlBuffer := strings.NewReader(`
[[metadata.dependencies]]
	id = "some-dependency"
	sha256 = "some-sha"
	source_sha256 = "some-source-sha"
	stacks = ["some-stack"]
	uri = "http://some-url"
	version = "1.2.3"
`)
					var config cargo.Config
					Expect(cargo.DecodeConfig(tomlBuffer, &config)).To(Succeed())

					Expect(config.Metadata.Dependencies).To(HaveLen(1))
					Expect(config.Metadata.Dependencies[0].SHA256).To(Equal("some-sha"))
					Expect(config.Metadata.Dependencies[0].SourceSHA256).To(Equal("some-source-sha"))
					Expect(config.Metadata.Dependencies[0].Stacks).To(Equal([]string{"some-stack"}))
				})
			})

			context("when both the legacy and standard keys are present", func() {
				it("decodes both", func() {
					tomlBuffer := strings.NewReader(`
[[metadata.dependencies]]
	id = "some-dependency"
	eol-date = "2020-06-01T00:00:00Z"
	deprecation_date = "2019-01-01T00:00:00Z"
	uri = "http://some-url"
	version = "1.2.3"
`)
					eolDate, err := time.Parse(time.RFC3339, "2020-06-01T00:00:00Z")
					Expect(err).NotTo(HaveOccurred())

					deprecationDate, err := time.Parse(time.RFC3339, "2019-01-01T00:00:00Z")
					Expect(err).NotTo(HaveOccurred())

					var config cargo.Config
					Expect(cargo.DecodeConfig(tomlBuffer, &config)).To(Succeed())

					Expect(config.Metadata.Dependencies).To(HaveLen(1))
					Expect(config.Metadata.Dependencies[0].EOLDate).To(Equal(&eolDate))
					Expect(config.Metadata.Dependencies[0].DeprecationDate).To(Equal(&deprecationDate))
				})
			})
		})

		context("EncodeConfig", func() {
			context("when the dependency uses the legacy CPE and PURL fields", func() {
				it("encodes the legacy cpe and purl keys", func() {
					err := cargo.EncodeConfig(buffer, cargo.Config{
						Buildpack: cargo.ConfigBuildpack{
							ID: "some-buildpack-id",
						},
						Metadata: cargo.ConfigMetadata{
							Dependencies: []cargo.ConfigMetadataDependency{
								{
									CPE:     "some-cpe",
									ID:      "some-dependency",
									PURL:    "some-purl",
									URI:     "http://some-url",
									Version: "1.2.3",
								},
							},
						},
					})
					Expect(err).NotTo(HaveOccurred())
					Expect(buffer.String()).To(MatchTOML(`
[buildpack]
	id = "some-buildpack-id"

[[metadata.dependencies]]
	cpe = "some-cpe"
	id = "some-dependency"
	purl = "some-purl"
	uri = "http://some-url"
	version = "1.2.3"
`))
				})
			})

			context("when a dependency was decoded from the legacy format", func() {
				it("writes the legacy keys back out unchanged", func() {
					tomlBuffer := strings.NewReader(`
[buildpack]
	id = "some-buildpack-id"

[[metadata.dependencies]]
	id = "some-dependency"
	cpe = "some-cpe"
	purl = "some-purl"
	deprecation_date = "2020-06-01T00:00:00Z"
	licenses = ["fancy-license"]
	sha256 = "some-sha"
	source-sha256 = "some-source-sha"
	stacks = ["some-stack"]
	uri = "http://some-url"
	version = "1.2.3"
`)
					var config cargo.Config
					Expect(cargo.DecodeConfig(tomlBuffer, &config)).To(Succeed())
					Expect(cargo.EncodeConfig(buffer, config)).To(Succeed())

					// Every legacy key is written back out in its original form so
					// that anything still reading the old format keeps working.
					Expect(buffer.String()).To(MatchTOML(`
[buildpack]
	id = "some-buildpack-id"

[[metadata.dependencies]]
	cpe = "some-cpe"
	deprecation_date = "2020-06-01T00:00:00Z"
	id = "some-dependency"
	licenses = ["fancy-license"]
	purl = "some-purl"
	sha256 = "some-sha"
	source-sha256 = "some-source-sha"
	stacks = ["some-stack"]
	uri = "http://some-url"
	version = "1.2.3"
`))
				})
			})

			context("when the dependency uses the old source_sha256 spelling", func() {
				it("still reads it", func() {
					tomlBuffer := strings.NewReader(`
[[metadata.dependencies]]
	id = "some-dependency"
	source_sha256 = "some-source-sha"
	uri = "http://some-url"
	version = "1.2.3"
`)
					var config cargo.Config
					Expect(cargo.DecodeConfig(tomlBuffer, &config)).To(Succeed())

					Expect(config.Metadata.Dependencies).To(HaveLen(1))
					Expect(config.Metadata.Dependencies[0].SourceSHA256).To(Equal("some-source-sha"))
				})
			})

			context("when a dependency license has no type", func() {
				it("does not add an empty type", func() {
					tomlBuffer := strings.NewReader(`
[buildpack]
	id = "some-buildpack-id"

[[metadata.dependencies]]
	id = "some-dependency"
	uri = "http://some-url"
	version = "1.2.3"

	[[metadata.dependencies.licenses]]
	uri = "http://some-url/license"
`)
					var config cargo.Config
					Expect(cargo.DecodeConfig(tomlBuffer, &config)).To(Succeed())
					Expect(cargo.EncodeConfig(buffer, config)).To(Succeed())

					Expect(buffer.String()).To(MatchTOML(`
[buildpack]
	id = "some-buildpack-id"

[[metadata.dependencies]]
	id = "some-dependency"
	uri = "http://some-url"
	version = "1.2.3"

	[[metadata.dependencies.licenses]]
	uri = "http://some-url/license"
`))
				})
			})

			context("when a dependency was decoded from the standard format", func() {
				it("writes the standard keys back out unchanged", func() {
					tomlBuffer := strings.NewReader(`
[buildpack]
	id = "some-buildpack-id"

[[metadata.dependencies]]
	id = "some-dependency"
	cpes = ["some-cpe"]
	purls = ["some-purl"]
	eol-date = "2020-06-01T00:00:00Z"
	uri = "http://some-url"
	version = "1.2.3"

	[[metadata.dependencies.licenses]]
	type = "fancy-license"
	uri = "http://some-url/license"
`)
					var config cargo.Config
					Expect(cargo.DecodeConfig(tomlBuffer, &config)).To(Succeed())
					Expect(cargo.EncodeConfig(buffer, config)).To(Succeed())

					Expect(buffer.String()).To(MatchTOML(`
[buildpack]
	id = "some-buildpack-id"

[[metadata.dependencies]]
	cpes = ["some-cpe"]
	eol-date = "2020-06-01T00:00:00Z"
	id = "some-dependency"
	purls = ["some-purl"]
	uri = "http://some-url"
	version = "1.2.3"

	[[metadata.dependencies.licenses]]
	type = "fancy-license"
	uri = "http://some-url/license"
`))
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
