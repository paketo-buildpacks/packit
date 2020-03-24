package internal_test

import (
	"archive/tar"
	"compress/gzip"
	"io/ioutil"
	"os"
	"testing"

	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/cargo/jam/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuildpackInspector(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buildpack string
		inspector internal.BuildpackInspector
	)

	it.Before(func() {
		file, err := ioutil.TempFile("", "buildpack")
		Expect(err).NotTo(HaveOccurred())

		gr := gzip.NewWriter(file)
		tr := tar.NewWriter(gr)

		content := []byte(`[buildpack]
id = "some-buildpack"

[metadata.default-versions]
	some-dependency = "1.2.x"
	other-dependency = "2.3.x"

[[metadata.dependencies]]
	id = "some-dependency"
	stacks = ["some-stack"]
	version = "some-version"

[[metadata.dependencies]]
	id = "other-dependency"
	stacks = ["other-stack"]
	version = "other-version"
`)

		err = tr.WriteHeader(&tar.Header{
			Name: "./buildpack.toml",
			Mode: 0644,
			Size: int64(len(content)),
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = tr.Write(content)
		Expect(err).NotTo(HaveOccurred())

		Expect(tr.Close()).To(Succeed())
		Expect(gr.Close()).To(Succeed())

		buildpack = file.Name()

		Expect(file.Close()).To(Succeed())

		inspector = internal.NewBuildpackInspector()
	})

	it.After(func() {
		Expect(os.Remove(buildpack)).To(Succeed())
	})

	context("Dependencies", func() {
		it("returns a list of dependencies", func() {
			dependencies, defaultVersions, err := inspector.Dependencies(buildpack)
			Expect(err).NotTo(HaveOccurred())
			Expect(dependencies).To(Equal([]cargo.ConfigMetadataDependency{
				{
					ID:      "some-dependency",
					Stacks:  []string{"some-stack"},
					Version: "some-version",
				},
				{
					ID:      "other-dependency",
					Stacks:  []string{"other-stack"},
					Version: "other-version",
				},
			}))

			Expect(defaultVersions).To(Equal(map[string]string{
				"some-dependency":  "1.2.x",
				"other-dependency": "2.3.x",
			}))
		})

		context("failure cases", func() {
			context("when the tarball does not exist", func() {
				it("returns an error", func() {
					_, _, err := inspector.Dependencies("/tmp/no-such-file")
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			context("when the tarball is garbage", func() {
				it.Before(func() {
					err := ioutil.WriteFile(buildpack, []byte("%%%"), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				it("returns an error", func() {
					_, _, err := inspector.Dependencies(buildpack)
					Expect(err).To(MatchError(ContainSubstring("failed to open gzip reader")))
					Expect(err).To(MatchError(ContainSubstring("EOF")))
				})
			})

			context("when the toml inside of the tarball is bad", func() {
				it.Before(func() {
					file, err := ioutil.TempFile("", "buildpack")
					Expect(err).NotTo(HaveOccurred())

					gr := gzip.NewWriter(file)
					tr := tar.NewWriter(gr)

					content := []byte(`%%%`)

					err = tr.WriteHeader(&tar.Header{
						Name: "./buildpack.toml",
						Mode: 0644,
						Size: int64(len(content)),
					})
					Expect(err).NotTo(HaveOccurred())

					_, err = tr.Write(content)
					Expect(err).NotTo(HaveOccurred())

					Expect(tr.Close()).To(Succeed())
					Expect(gr.Close()).To(Succeed())

					buildpack = file.Name()

					Expect(file.Close()).To(Succeed())
				})

				it("returns an error", func() {
					_, _, err := inspector.Dependencies(buildpack)
					Expect(err).To(MatchError(ContainSubstring("bare keys cannot contain '%'")))
				})
			})

			context("when there is no buildpack.toml in the tarball", func() {
				it.Before(func() {
					file, err := ioutil.TempFile("", "buildpack")
					Expect(err).NotTo(HaveOccurred())

					gr := gzip.NewWriter(file)
					tr := tar.NewWriter(gr)

					content := []byte(`some contents`)

					err = tr.WriteHeader(&tar.Header{
						Name: "./some-file",
						Mode: 0644,
						Size: int64(len(content)),
					})
					Expect(err).NotTo(HaveOccurred())

					_, err = tr.Write(content)
					Expect(err).NotTo(HaveOccurred())

					Expect(tr.Close()).To(Succeed())
					Expect(gr.Close()).To(Succeed())

					buildpack = file.Name()

					Expect(file.Close()).To(Succeed())
				})

				it("returns an error", func() {
					_, _, err := inspector.Dependencies(buildpack)
					Expect(err).To(MatchError(ContainSubstring("failed to find buildpack.toml in buildpack tarball")))
				})
			})
		})
	})
}
