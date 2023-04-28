package cargo_test

import (
	"strings"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testExtensionConfig(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

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

[[metadata.some-map]]
	key = "value"

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
