package fs_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/packit/fs"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testMove(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("Move", func() {
		var (
			sourceDir      string
			destinationDir string
		)

		it.Before(func() {
			var err error
			sourceDir, err = ioutil.TempDir("", "source")
			Expect(err).NotTo(HaveOccurred())

			destinationDir, err = ioutil.TempDir("", "destination")
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(os.RemoveAll(sourceDir)).To(Succeed())
			Expect(os.RemoveAll(destinationDir)).To(Succeed())
		})

		context("when the source is a file", func() {
			var source, destination string

			it.Before(func() {
				source = filepath.Join(sourceDir, "source")
				destination = filepath.Join(destinationDir, "destination")

				err := ioutil.WriteFile(source, []byte("some-content"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("moves the source file to the destination", func() {
				err := fs.Move(source, destination)
				Expect(err).NotTo(HaveOccurred())

				content, err := ioutil.ReadFile(destination)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("some-content"))

				info, err := os.Stat(destination)
				Expect(err).NotTo(HaveOccurred())
				Expect(info.Mode()).To(Equal(os.FileMode(0644)))

				Expect(source).NotTo(BeAnExistingFile())
			})

			context("failure cases", func() {
				context("when the source cannot be read", func() {
					it.Before(func() {
						Expect(os.Chmod(source, 0000)).To(Succeed())
					})

					it("returns an error", func() {
						err := fs.Move(source, destination)
						Expect(err).To(MatchError(ContainSubstring("permission denied")))
					})
				})

				context("when the destination cannot be removed", func() {
					it.Before(func() {
						Expect(os.Chmod(destinationDir, 0000)).To(Succeed())
					})

					it("returns an error", func() {
						err := fs.Move(source, destination)
						Expect(err).To(MatchError(ContainSubstring("failed to move: destination exists:")))
						Expect(err).To(MatchError(ContainSubstring("permission denied")))
					})
				})
			})
		})

		context("when the source is a directory", func() {
			var source, destination, external string

			it.Before(func() {
				var err error
				external, err = ioutil.TempDir("", "external")
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(filepath.Join(external, "some-file"), []byte("some-content"), 0644)
				Expect(err).NotTo(HaveOccurred())

				source = filepath.Join(sourceDir, "source")
				destination = filepath.Join(destinationDir, "destination")

				Expect(os.MkdirAll(filepath.Join(source, "some-dir"), os.ModePerm)).To(Succeed())

				err = ioutil.WriteFile(filepath.Join(source, "some-dir", "some-file"), []byte("some-content"), 0644)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(filepath.Join(source, "some-dir", "readonly-file"), []byte("some-content"), 0444)
				Expect(err).NotTo(HaveOccurred())

				err = os.Symlink(filepath.Join(source, "some-dir", "some-file"), filepath.Join(source, "some-dir", "some-symlink"))
				Expect(err).NotTo(HaveOccurred())

				err = os.Symlink(filepath.Join(external, "some-file"), filepath.Join(source, "some-dir", "external-symlink"))
				Expect(err).NotTo(HaveOccurred())
			})

			context("when the destination does not exist", func() {
				it("moves the source directory to the destination", func() {
					err := fs.Move(source, destination)
					Expect(err).NotTo(HaveOccurred())

					Expect(destination).To(BeADirectory())

					content, err := ioutil.ReadFile(filepath.Join(destination, "some-dir", "some-file"))
					Expect(err).NotTo(HaveOccurred())
					Expect(string(content)).To(Equal("some-content"))

					info, err := os.Stat(filepath.Join(destination, "some-dir", "some-file"))
					Expect(err).NotTo(HaveOccurred())
					Expect(info.Mode()).To(Equal(os.FileMode(0644)))

					info, err = os.Stat(filepath.Join(destination, "some-dir", "readonly-file"))
					Expect(err).NotTo(HaveOccurred())
					Expect(info.Mode()).To(Equal(os.FileMode(0444)))

					path, err := os.Readlink(filepath.Join(destination, "some-dir", "some-symlink"))
					Expect(err).NotTo(HaveOccurred())
					Expect(path).To(Equal(filepath.Join(destination, "some-dir", "some-file")))

					path, err = os.Readlink(filepath.Join(destination, "some-dir", "external-symlink"))
					Expect(err).NotTo(HaveOccurred())
					Expect(path).To(Equal(filepath.Join(external, "some-file")))

					Expect(source).NotTo(BeAnExistingFile())
				})
			})

			context("when the destination is a file", func() {
				it.Before(func() {
					Expect(os.RemoveAll(destination))
					Expect(ioutil.WriteFile(destination, []byte{}, os.ModePerm)).To(Succeed())
				})

				it("moves the source directory to the destination", func() {
					err := fs.Move(source, destination)
					Expect(err).NotTo(HaveOccurred())

					Expect(destination).To(BeADirectory())

					content, err := ioutil.ReadFile(filepath.Join(destination, "some-dir", "some-file"))
					Expect(err).NotTo(HaveOccurred())
					Expect(string(content)).To(Equal("some-content"))

					info, err := os.Stat(filepath.Join(destination, "some-dir", "some-file"))
					Expect(err).NotTo(HaveOccurred())
					Expect(info.Mode()).To(Equal(os.FileMode(0644)))

					info, err = os.Stat(filepath.Join(destination, "some-dir", "readonly-file"))
					Expect(err).NotTo(HaveOccurred())
					Expect(info.Mode()).To(Equal(os.FileMode(0444)))

					path, err := os.Readlink(filepath.Join(destination, "some-dir", "some-symlink"))
					Expect(err).NotTo(HaveOccurred())
					Expect(path).To(Equal(filepath.Join(destination, "some-dir", "some-file")))

					path, err = os.Readlink(filepath.Join(destination, "some-dir", "external-symlink"))
					Expect(err).NotTo(HaveOccurred())
					Expect(path).To(Equal(filepath.Join(external, "some-file")))

					Expect(source).NotTo(BeAnExistingFile())
				})
			})

			context("failure cases", func() {
				context("when the source does not exist", func() {
					it("returns an error", func() {
						err := fs.Move("no-such-source", destination)
						Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
					})
				})

				context("when the source cannot be walked", func() {
					it.Before(func() {
						Expect(os.Chmod(source, 0000)).To(Succeed())
					})

					it.After(func() {
						Expect(os.Chmod(source, 0777)).To(Succeed())
					})

					it("returns an error", func() {
						err := fs.Move(source, destination)
						Expect(err).To(MatchError(ContainSubstring("permission denied")))
					})
				})
			})
		})
	})
}
