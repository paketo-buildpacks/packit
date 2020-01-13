package cargo_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/packit/cargo"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/packit/matchers"
	. "github.com/onsi/gomega"
)

func testFileBundler(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		fileBundler cargo.FileBundler
	)

	it.Before(func() {
		fileBundler = cargo.NewFileBundler()
	})

	context("Bundle", func() {
		it("returns a list of cargo files", func() {
			files, err := fileBundler.Bundle(filepath.Join("jam", "testdata", "example-cnb"), []string{"bin/build", "bin/detect", "buildpack.toml"}, cargo.Config{
				API: "0.2",
				Buildpack: cargo.ConfigBuildpack{
					ID:      "other-buildpack-id",
					Name:    "other-buildpack-name",
					Version: "other-buildpack-version",
				},
				Metadata: cargo.ConfigMetadata{
					IncludeFiles: []string{
						"bin/build",
						"bin/detect",
						"buildpack.toml",
					},
					PrePackage: "some-pre-package-script.sh",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(files).To(HaveLen(3))

			Expect(files[0].Name).To(Equal("bin/build"))
			Expect(files[0].Info.Size()).To(Equal(int64(14)))
			Expect(files[0].Info.Mode()).To(Equal(os.FileMode(0755)))

			content, err := ioutil.ReadAll(files[0])
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("build-contents"))

			Expect(files[1].Name).To(Equal("bin/detect"))
			Expect(files[1].Info.Size()).To(Equal(int64(15)))
			Expect(files[1].Info.Mode()).To(Equal(os.FileMode(0755)))

			content, err = ioutil.ReadAll(files[1])
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("detect-contents"))

			Expect(files[2].Name).To(Equal("buildpack.toml"))
			Expect(files[2].Info.Size()).To(Equal(int64(244)))
			Expect(files[2].Info.Mode()).To(Equal(os.FileMode(0644)))

			content, err = ioutil.ReadAll(files[2])
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(MatchTOML(`api = "0.2"
[buildpack]
id = "other-buildpack-id"
name = "other-buildpack-name"
version = "other-buildpack-version"

[metadata]
include_files = ["bin/build", "bin/detect", "buildpack.toml"]
pre_package = "some-pre-package-script.sh"`))
		})

		context("error cases", func() {
			context("when included file does not exist", func() {
				it("fails", func() {
					_, err := fileBundler.Bundle(filepath.Join("jam", "testdata", "example-cnb"), []string{"bin/fake/build", "bin/detect", "buildpack.toml"}, cargo.Config{})
					Expect(err).To(MatchError(ContainSubstring("error opening included file:")))
				})
			})
		})
	})
}
