package cargo_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/cloudfoundry/packit/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuildpackParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path   string
		parser cargo.BuildpackParser
	)

	it.Before(func() {
		file, err := ioutil.TempFile("", "buildpack.toml")
		Expect(err).NotTo(HaveOccurred())

		_, err = file.WriteString(`api = "0.2"
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
`)
		Expect(err).NotTo(HaveOccurred())

		Expect(file.Close()).To(Succeed())

		path = file.Name()

		parser = cargo.NewBuildpackParser()
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("Parse", func() {
		it("parses a given buildpack.toml", func() {
			config, err := parser.Parse(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(cargo.Config{
				API: "0.2",
				Buildpack: cargo.ConfigBuildpack{
					ID:      "some-buildpack-id",
					Name:    "some-buildpack-name",
					Version: "some-buildpack-version",
				},
				Metadata: cargo.ConfigMetadata{
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

		context("when the buildpack.toml does not exist", func() {
			it.Before(func() {
				Expect(os.Remove(path)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := parser.Parse(path)
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})
		})

		context("when the buildpack.toml is malformed", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte("%%%"), 0644)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := parser.Parse(path)
				Expect(err).To(MatchError(ContainSubstring("keys cannot contain '%'")))
			})
		})
	})
}
