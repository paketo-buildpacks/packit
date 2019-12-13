package cargo_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testTarBuilder(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		tempFile string
		tempDir  string
		output   *bytes.Buffer
		builder  cargo.TarBuilder
	)

	it.Before(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "output")
		Expect(err).ToNot(HaveOccurred())

		tempFile = filepath.Join(tempDir, "buildpack.tgz")

		output = bytes.NewBuffer(nil)
		builder = cargo.NewTarBuilder(scribe.NewLogger(output))
	})

	it.After(func() {
		Expect(os.RemoveAll(tempDir)).To(Succeed())
	})

	context("Build", func() {
		context("given a destination and a list of files", func() {
			it("constructs a tar ball", func() {
				err := builder.Build(tempFile, []cargo.File{
					{
						Name:       "bin/build",
						Size:       int64(len("build-contents")),
						Mode:       int64(0755),
						ReadCloser: ioutil.NopCloser(strings.NewReader("build-contents")),
					},
					{
						Name:       "bin/detect",
						Size:       int64(len("detect-contents")),
						Mode:       int64(0755),
						ReadCloser: ioutil.NopCloser(strings.NewReader("detect-contents")),
					},
					{
						Name:       "buildpack.toml",
						Size:       int64(len("buildpack-toml-contents")),
						Mode:       int64(0644),
						ReadCloser: ioutil.NopCloser(strings.NewReader("buildpack-toml-contents")),
					},
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(output.String()).To(ContainSubstring(fmt.Sprintf("Building tarball: %s", tempFile)))
				Expect(output.String()).To(ContainSubstring("bin/build"))
				Expect(output.String()).To(ContainSubstring("bin/detect"))
				Expect(output.String()).To(ContainSubstring("buildpack.toml"))

				file, err := os.Open(tempFile)
				Expect(err).ToNot(HaveOccurred())

				contents, hdr, err := ExtractFile(file, "buildpack.toml")
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal("buildpack-toml-contents"))
				Expect(hdr.Mode).To(Equal(int64(0644)))

				contents, hdr, err = ExtractFile(file, "bin/build")
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal("build-contents"))
				Expect(hdr.Mode).To(Equal(int64(0755)))

				contents, hdr, err = ExtractFile(file, "bin/detect")
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(Equal("detect-contents"))
				Expect(hdr.Mode).To(Equal(int64(0755)))
			})
		})

		context("failure cases", func() {
			context("when it is unable to create the destination file", func() {
				it.Before(func() {
					Expect(os.Chmod(tempDir, 0000)).To(Succeed())
				})

				it.Before(func() {
					Expect(os.Chmod(tempDir, 0644)).To(Succeed())
				})

				it("returns an error", func() {
					err := builder.Build(tempFile, []cargo.File{
						{
							Name:       "bin/build",
							Size:       int64(len("build-contents")),
							ReadCloser: ioutil.NopCloser(strings.NewReader("build-contents")),
						},
					})
					Expect(err).To(MatchError(ContainSubstring("failed to create tarball")))
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when one of the files cannot be written", func() {
				it("returns an error", func() {
					err := builder.Build(tempFile, []cargo.File{
						{
							Name:       "bin/build",
							ReadCloser: ioutil.NopCloser(strings.NewReader("build-contents")),
						},
					})
					Expect(err).To(MatchError(ContainSubstring("failed to write file to tarball")))
					Expect(err).To(MatchError(ContainSubstring("write too long")))
				})
			})
		})
	})
}
