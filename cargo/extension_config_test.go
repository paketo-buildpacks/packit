package cargo_test

import (
	"bytes"

	"strings"
	"testing"

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

			err := cargo.EncodeExtensionConfig(buffer, cargo.ExtensionConfig{
				API: "0.7",
				Extension: cargo.ConfigExtension{
					ID:          "some-extension-id",
					Name:        "some-extension-name",
					Version:     "some-extension-version",
					Homepage:    "some-extension-homepage",
					Description: "some-extension-description",
					Keywords:    []string{"some-extension-keyword"},
					Licenses: []cargo.ConfigExtensionLicense{
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
    licenses = ["fancy-license", "fancy-license-2"]
    name = "Some Dependency"
    sha256 = "shasum"
    source = "source"
    source-checksum = "sha256:source-shasum"
    source_sha256 = "source-shasum"
    stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
    uri = "http://some-url"
    version = "1.2.3"
`))
		})

		it("encodes the config to TOML when the config dependency licenses are structured like ConfigExtensionLicenses ", func() {

			err := cargo.EncodeExtensionConfig(buffer, cargo.ExtensionConfig{
				API: "0.7",
				Extension: cargo.ConfigExtension{
					ID:          "some-extension-id",
					Name:        "some-extension-name",
					Version:     "some-extension-version",
					Homepage:    "some-extension-homepage",
					Description: "some-extension-description",
					Keywords:    []string{"some-extension-keyword"},
					Licenses: []cargo.ConfigExtensionLicense{
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
							Licenses: []interface{}{
								cargo.ConfigBuildpackLicense{
									Type: "fancy-license",
									URI:  "some-license-uri",
								},
								cargo.ConfigBuildpackLicense{
									Type: "fancy-license-2",
									URI:  "some-license-uri",
								},
							}, Name: "Some Dependency",
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
`))
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
					Licenses: []cargo.ConfigExtensionLicense{
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
	})

}
