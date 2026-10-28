package cargo_test

import (
	"bytes"

	"strings"
	"testing"
	"time"

	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/packit/v2/matchers"
)

func testExtensionConfig(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		buffer *bytes.Buffer
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
	})

	context("EncodeExtensionConfig", func() {
		it("encodes the extension config to TOML", func() {
			eolDate, err := time.Parse(time.RFC3339, "2020-06-01T00:00:00Z")
			Expect(err).NotTo(HaveOccurred())

			err = cargo.EncodeExtensionConfig(buffer, cargo.ExtensionConfig{
				API: "0.7",
				Extension: cargo.ConfigExtension{
					ID:          "some-extension-id",
					Name:        "some-extension-name",
					Version:     "some-extension-version",
					Homepage:    "some-extension-homepage",
					Description: "some-extension-description",
					Keywords:    []string{"some-extension-keyword"},
					Licenses: []cargo.ConfigBuildpackLicense{
						{
							Type: "some-license-type",
							URI:  "some-license-uri",
						},
					},
				},
				Metadata: cargo.ConfigExtensionMetadata{
					IncludeFiles: []string{
						"some-include-file",
						"other-include-file",
					},
					PrePackage: "some-pre-package-script.sh",
					Dependencies: []cargo.ConfigExtensionMetadataDependency{
						{
							Arch:            "some-arch",
							Checksum:        "sha256:some-sum",
							CPE:             "some-cpe",
							CPEs:            []string{"some-cpe", "other-cpe"},
							EOLDate:         &eolDate,
							ID:              "some-dependency",
							Licenses:        []interface{}{"fancy-license", "fancy-license-2"},
							Name:            "Some Dependency",
							OS:              "some-os",
							PURL:            "some-purl",
							PURLs:           []string{"some-purl", "other-purl"},
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
					Configurations: []cargo.ConfigExtensionMetadataConfiguration{
						{
							Default:     "0",
							Description: "some-metadata-configuration-description",
							Launch:      true,
							Name:        "SOME_METADATA_CONFIGURATION_NAME",
							Build:       true,
						},
					},

					DefaultVersions: map[string]string{
						"some-dependency": "1.2.x",
					},
				},
				Targets: []cargo.ConfigTarget{
					{
						OS:   "linux",
						Arch: "arm64",
					},
					{
						OS:   "linux",
						Arch: "amd64",
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(buffer.String()).To(MatchTOML(`
api = "0.7"

[extension]
  description = "some-extension-description"
  homepage = "some-extension-homepage"
  id = "some-extension-id"
  keywords = ["some-extension-keyword"]
  name = "some-extension-name"
  version = "some-extension-version"

  [[extension.licenses]]
    type = "some-license-type"
    uri = "some-license-uri"

[metadata]
  include-files = ["some-include-file", "other-include-file"]
  pre-package = "some-pre-package-script.sh"

  [[metadata.configurations]]
    build = true
    default = "0"
    description = "some-metadata-configuration-description"
    launch = true
    name = "SOME_METADATA_CONFIGURATION_NAME"
  [metadata.default-versions]
    some-dependency = "1.2.x"

  [[metadata.dependencies]]
    arch = "some-arch"
    checksum = "sha256:some-sum"
    cpe = "some-cpe"
    cpes = ["some-cpe", "other-cpe"]
    eol-date = "2020-06-01T00:00:00Z"
    id = "some-dependency"
    licenses = ["fancy-license", "fancy-license-2"]
    name = "Some Dependency"
    os = "some-os"
    purl = "some-purl"
    purls = ["some-purl", "other-purl"]
    sha256 = "shasum"
    source = "source"
    source-checksum = "sha256:source-shasum"
    source-sha256 = "source-shasum"
    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
    strip-components = 1
    uri = "http://some-url"
    version = "1.2.3"

[[targets]]
  os = "linux"
  arch = "arm64"

[[targets]]
  os = "linux"
  arch = "amd64"
`))
		})

		context("when the config dependency licenses are structured like ConfigDependencyLicense", func() {
			it("encodes the config to TOML", func() {

				err := cargo.EncodeExtensionConfig(buffer, cargo.ExtensionConfig{
					API: "0.7",
					Extension: cargo.ConfigExtension{
						ID:          "some-extension-id",
						Name:        "some-extension-name",
						Version:     "some-extension-version",
						Homepage:    "some-extension-homepage",
						Description: "some-extension-description",
						Keywords:    []string{"some-extension-keyword"},
						Licenses: []cargo.ConfigBuildpackLicense{
							{
								Type: "some-license-type",
								URI:  "some-license-uri",
							},
						},
					},
					Metadata: cargo.ConfigExtensionMetadata{
						IncludeFiles: []string{
							"some-include-file",
							"other-include-file",
						},
						PrePackage: "some-pre-package-script.sh",
						Dependencies: []cargo.ConfigExtensionMetadataDependency{
							{
								Checksum: "sha256:some-sum",
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
						Configurations: []cargo.ConfigExtensionMetadataConfiguration{
							{
								Default:     "0",
								Description: "some-metadata-configuration-description",
								Launch:      true,
								Name:        "SOME_METADATA_CONFIGURATION_NAME",
								Build:       true,
							},
						},
						DefaultVersions: map[string]string{
							"some-dependency": "1.2.x",
						},
					},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(buffer.String()).To(MatchTOML(`
api = "0.7"

[extension]
  description = "some-extension-description"
  homepage = "some-extension-homepage"
  id = "some-extension-id"
  keywords = ["some-extension-keyword"]
  name = "some-extension-name"
  version = "some-extension-version"

  [[extension.licenses]]
    type = "some-license-type"
    uri = "some-license-uri"

[metadata]
  include-files = ["some-include-file", "other-include-file"]
  pre-package = "some-pre-package-script.sh"

  [[metadata.configurations]]
    build = true
    default = "0"
    description = "some-metadata-configuration-description"
    launch = true
    name = "SOME_METADATA_CONFIGURATION_NAME"
  [metadata.default-versions]
    some-dependency = "1.2.x"

  [[metadata.dependencies]]
    checksum = "sha256:some-sum"
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
`))
			})
		})
	})

	context("DecodeExtensionConfig", func() {
		it("decodes TOML to Extensionconfig", func() {
			tomlBuffer := strings.NewReader(`
api = "0.7"

[extension]
id = "some-extension-id"
name = "some-extension-name"
version = "some-extension-version"
homepage = "some-extension-homepage"
description = "some-extension-description"
keywords = [ "some-extension-keyword" ]

[[extension.licenses]]
	type = "some-license-type"
	uri = "some-license-uri"

[metadata]
	include-files = ["some-include-file", "other-include-file"]
	pre-package = "some-pre-package-script.sh"

[metadata.default-versions]
	some-dependency = "1.2.x"

[[metadata.dependencies]]
	checksum = "sha256:some-sum"
	id = "some-dependency"
	licenses = ["fancy-license", "fancy-license-2"]
	name = "Some Dependency"
	sha256 = "shasum"
	source = "source"
	source-checksum = "sha256:source-shasum"
	source_sha256 = "source-shasum"
	stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
	uri = "http://some-url"
	version = "1.2.3"

[[metadata.configurations]]
	default = "0"
	description = "some-metadata-configuration-description"
	launch = true
	name = "SOME_METADATA_CONFIGURATION_NAME"
	build = true
`)

			var config cargo.ExtensionConfig
			Expect(cargo.DecodeExtensionConfig(tomlBuffer, &config)).To(Succeed())
			Expect(config).To(Equal(cargo.ExtensionConfig{
				API: "0.7",
				Extension: cargo.ConfigExtension{
					ID:          "some-extension-id",
					Name:        "some-extension-name",
					Version:     "some-extension-version",
					Homepage:    "some-extension-homepage",
					Description: "some-extension-description",
					Keywords:    []string{"some-extension-keyword"},
					Licenses: []cargo.ConfigBuildpackLicense{
						{
							Type: "some-license-type",
							URI:  "some-license-uri",
						},
					},
				},
				Metadata: cargo.ConfigExtensionMetadata{
					IncludeFiles: []string{
						"some-include-file",
						"other-include-file",
					},
					PrePackage: "some-pre-package-script.sh",
					Dependencies: []cargo.ConfigExtensionMetadataDependency{
						{
							Checksum:       "sha256:some-sum",
							ID:             "some-dependency",
							Licenses:       []interface{}{"fancy-license", "fancy-license-2"},
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
					Configurations: []cargo.ConfigExtensionMetadataConfiguration{
						{
							Default:     "0",
							Description: "some-metadata-configuration-description",
							Launch:      true,
							Name:        "SOME_METADATA_CONFIGURATION_NAME",
							Build:       true,
						},
					},
					DefaultVersions: map[string]string{
						"some-dependency": "1.2.x",
					},
				},
			}))
		})

		context("dependency license are not a list of IDs", func() {
			it("decodes TOML to extensionConfig", func() {
				tomlBuffer := strings.NewReader(`
api = "0.2"

[extension]
  id = "some-extension-id"
  name = "some-extension-name"
  version = "some-extension-version"
  homepage = "some-extension-homepage"

  [[extension.licenses]]
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
`)

				var config cargo.ExtensionConfig
				Expect(cargo.DecodeExtensionConfig(tomlBuffer, &config)).To(Succeed())
				Expect(config).To(Equal(cargo.ExtensionConfig{
					API: "0.2",
					Extension: cargo.ConfigExtension{
						ID:       "some-extension-id",
						Name:     "some-extension-name",
						Version:  "some-extension-version",
						Homepage: "some-extension-homepage",
						Licenses: []cargo.ConfigBuildpackLicense{
							{
								Type: "some-license-type",
								URI:  "some-license-uri",
							},
						},
					},
					Metadata: cargo.ConfigExtensionMetadata{
						IncludeFiles: []string{
							"some-include-file",
							"other-include-file",
						},
						PrePackage: "some-pre-package-script.sh",
						Dependencies: []cargo.ConfigExtensionMetadataDependency{
							{
								Checksum: "sha256:some-sum",
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
						DefaultVersions: map[string]string{
							"some-dependency": "1.2.x",
						},
					},
				}))
			})

		})

		context("failure cases", func() {
			context("when a bad reader is passed in", func() {
				it("returns an error", func() {
					err := cargo.DecodeExtensionConfig(errorReader{}, &cargo.ExtensionConfig{})
					Expect(err).To(MatchError(ContainSubstring("failed to read")))
				})
			})
		})
	})

	// The backwards compatibility contexts below cover the dependency metadata
	// keys that predate the standard dependency metadata format (RFC 0059).
	// They exist to guarantee that extension.toml files written in the older
	// format keep working. Remove them once the legacy keys are no longer
	// supported.
	context("backwards compatibility", func() {
		context("DecodeExtensionConfig", func() {
			it("decodes the legacy dependency keys", func() {
				tomlBuffer := strings.NewReader(`
api = "0.7"

[[metadata.dependencies]]
	id = "some-dependency"
	cpe = "some-cpe"
	purl = "some-purl"
	deprecation_date = "2020-06-01T00:00:00Z"
	licenses = ["fancy-license"]
	sha256 = "some-sha"
	source_sha256 = "some-source-sha"
	stacks = ["some-stack"]
	uri = "http://some-url"
	version = "1.2.3"
`)
				deprecationDate, err := time.Parse(time.RFC3339, "2020-06-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())

				var config cargo.ExtensionConfig
				Expect(cargo.DecodeExtensionConfig(tomlBuffer, &config)).To(Succeed())

				Expect(config.Metadata.Dependencies).To(HaveLen(1))

				dependency := config.Metadata.Dependencies[0]
				Expect(dependency.CPE).To(Equal("some-cpe"))
				Expect(dependency.CPEs).To(BeEmpty())
				Expect(dependency.PURL).To(Equal("some-purl"))
				Expect(dependency.PURLs).To(BeEmpty())
				Expect(dependency.DeprecationDate).To(Equal(&deprecationDate))
				Expect(dependency.Licenses).To(Equal([]interface{}{"fancy-license"}))
				Expect(dependency.LicenseEntries).To(BeEmpty())
				Expect(dependency.SHA256).To(Equal("some-sha"))
				Expect(dependency.SourceSHA256).To(Equal("some-source-sha"))
				Expect(dependency.Stacks).To(Equal([]string{"some-stack"}))
			})

			it("decodes the standard dependency keys", func() {
				tomlBuffer := strings.NewReader(`
api = "0.7"

[[metadata.dependencies]]
	id = "some-dependency"
	cpes = ["some-cpe"]
	purls = ["some-purl"]
	eol-date = "2020-06-01T00:00:00Z"
	os = "linux"
	arch = "amd64"
	source-sha256 = "some-source-sha"
	uri = "http://some-url"
	version = "1.2.3"

	[[metadata.dependencies.distros]]
	name = "ubuntu"
	version = "22.04"

	[[metadata.dependencies.licenses]]
	type = "fancy-license"
	uri = "http://some-url/license"
`)
				eolDate, err := time.Parse(time.RFC3339, "2020-06-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())

				var config cargo.ExtensionConfig
				Expect(cargo.DecodeExtensionConfig(tomlBuffer, &config)).To(Succeed())

				Expect(config.Metadata.Dependencies).To(HaveLen(1))

				dependency := config.Metadata.Dependencies[0]
				Expect(dependency.CPEs).To(Equal([]string{"some-cpe"}))
				Expect(dependency.PURLs).To(Equal([]string{"some-purl"}))
				Expect(dependency.EOLDate).To(Equal(&eolDate))
				Expect(dependency.OS).To(Equal("linux"))
				Expect(dependency.Arch).To(Equal("amd64"))
				Expect(dependency.SourceSHA256).To(Equal("some-source-sha"))
				Expect(dependency.Distros).To(Equal([]cargo.ConfigDistro{{Name: "ubuntu", Version: "22.04"}}))
				Expect(dependency.LicenseEntries).To(Equal([]cargo.ConfigDependencyLicense{
					{Type: "fancy-license", URI: "http://some-url/license"},
				}))
			})
		})

		context("EncodeExtensionConfig", func() {
			it("writes the legacy keys back out unchanged", func() {
				tomlBuffer := strings.NewReader(`
api = "0.7"

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
				var config cargo.ExtensionConfig
				Expect(cargo.DecodeExtensionConfig(tomlBuffer, &config)).To(Succeed())
				Expect(cargo.EncodeExtensionConfig(buffer, config)).To(Succeed())

				Expect(buffer.String()).To(MatchTOML(`
api = "0.7"

[extension]

[metadata]

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

			it("does not add an empty license type", func() {
				tomlBuffer := strings.NewReader(`
api = "0.7"

[[metadata.dependencies]]
	id = "some-dependency"
	uri = "http://some-url"
	version = "1.2.3"

	[[metadata.dependencies.licenses]]
	uri = "http://some-url/license"
`)
				var config cargo.ExtensionConfig
				Expect(cargo.DecodeExtensionConfig(tomlBuffer, &config)).To(Succeed())
				Expect(cargo.EncodeExtensionConfig(buffer, config)).To(Succeed())

				Expect(buffer.String()).To(MatchTOML(`
api = "0.7"

[extension]

[metadata]

[[metadata.dependencies]]
	id = "some-dependency"
	uri = "http://some-url"
	version = "1.2.3"

	[[metadata.dependencies.licenses]]
	uri = "http://some-url/license"
`))
			})
		})
	})
}
