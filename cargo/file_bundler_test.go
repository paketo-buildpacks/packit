package cargo_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/packit/matchers"
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
			files, err := fileBundler.Bundle(filepath.Join("jam", "testdata", "example-cnb"), []string{"bin/build", "bin/detect", "bin/link", "buildpack.toml"}, cargo.Config{
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
						"bin/link",
						"buildpack.toml",
					},
					PrePackage: "some-pre-package-script.sh",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(files).To(HaveLen(4))

			Expect(files[0].Name).To(Equal("bin/build"))
			Expect(files[0].Info.Size()).To(Equal(int64(14)))
			Expect(files[0].Info.Mode()).To(Equal(os.FileMode(0755)))
			Expect(files[0].Link).To(Equal(""))

			content, err := ioutil.ReadAll(files[0])
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("build-contents"))

			Expect(files[1].Name).To(Equal("bin/detect"))
			Expect(files[1].Info.Size()).To(Equal(int64(15)))
			Expect(files[1].Info.Mode()).To(Equal(os.FileMode(0755)))
			Expect(files[1].Link).To(Equal(""))

			content, err = ioutil.ReadAll(files[1])
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("detect-contents"))

			Expect(files[2].Name).To(Equal("bin/link"))
			Expect(files[2].Info.Size()).To(Equal(int64(7)))
			Expect(files[2].Info.Mode() & os.ModeSymlink).To(Equal(os.ModeSymlink))
			Expect(files[2].Link).To(Equal("build"))
			Expect(files[2].ReadCloser).To(BeNil())

			Expect(files[3].Name).To(Equal("buildpack.toml"))
			Expect(files[3].Info.Size()).To(Equal(int64(256)))
			Expect(files[3].Info.Mode()).To(Equal(os.FileMode(0644)))
			Expect(files[3].Link).To(Equal(""))

			content, err = ioutil.ReadAll(files[3])
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(MatchTOML(`api = "0.2"
[buildpack]
id = "other-buildpack-id"
name = "other-buildpack-name"
version = "other-buildpack-version"

[metadata]
include-files = ["bin/build", "bin/detect", "bin/link", "buildpack.toml"]
pre-package = "some-pre-package-script.sh"`))
		})

		context("error cases", func() {
			context("when included file does not exist", func() {
				it("fails", func() {
					_, err := fileBundler.Bundle(filepath.Join("jam", "testdata", "example-cnb"), []string{"bin/fake/build", "bin/detect", "buildpack.toml"}, cargo.Config{})
					Expect(err).To(MatchError(ContainSubstring("error stating included file:")))
				})
			})
		})
	})
}
