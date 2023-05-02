package cargo_test

import (
	"os"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testExtensionParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path   string
		parser cargo.ExtensionParser
	)

	it.Before(func() {
		file, err := os.CreateTemp("", "extension.toml")
		Expect(err).NotTo(HaveOccurred())

		_, err = file.WriteString(`api = "0.7"
[extension]
id = "some-extension-id"
name = "some-extension-name"
version = "some-extension-version"

[metadata]
	include-files = ["some-include-file", "other-include-file"]
	pre-package = "some-pre-package-script.sh"

[[metadata.some-map]]
	key = "value"

[[metadata.dependencies]]
	id = "some-dependency"
	name = "Some Dependency"
	sha256 = "shasum"
	source = "source"
	source_sha256 = "source-shasum"
	stacks = ["io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"]
	uri = "http://some-url"
	version = "1.2.3"
`)
		Expect(err).NotTo(HaveOccurred())

		Expect(file.Close()).To(Succeed())

		path = file.Name()

		parser = cargo.NewExtensionParser()
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("Parse", func() {
		it("parses a given extension.toml", func() {
			config, err := parser.Parse(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(cargo.ExtensionConfig{
				API: "0.7",
				Extension: cargo.ConfigExtension{
					ID:      "some-extension-id",
					Name:    "some-extension-name",
					Version: "some-extension-version",
				},
				Metadata: cargo.ConfigExtensionMetadata{
					IncludeFiles: []string{
						"some-include-file",
						"other-include-file",
					},
					PrePackage: "some-pre-package-script.sh",
					Dependencies: []cargo.ConfigExtensionMetadataDependency{
						{
							ID:           "some-dependency",
							Name:         "Some Dependency",
							SHA256:       "shasum",
							Source:       "source",
							SourceSHA256: "source-shasum",
							Stacks:       []string{"io.buildpacks.stacks.bionic", "org.cloudfoundry.stacks.tiny"},
							URI:          "http://some-url",
							Version:      "1.2.3",
						},
					},
				},
			}))
		})

		context("when the extension.toml does not exist", func() {
			it.Before(func() {
				Expect(os.Remove(path)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := parser.Parse(path)
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})
		})

		context("when the extension.toml is malformed", func() {
			it.Before(func() {
				Expect(os.WriteFile(path, []byte("%%%"), 0644)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := parser.Parse(path)
				Expect(err).To(MatchError(ContainSubstring("expected '.' or '=', but got '%' instead")))
			})
		})
	})
}
