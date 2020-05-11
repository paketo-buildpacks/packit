package cargo_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDirectoryDuplicator(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect       = NewWithT(t).Expect
		destDir      string
		sourceDir    string
		directoryDup cargo.DirectoryDuplicator
	)

	it.Before(func() {
		var err error

		sourceDir, err = ioutil.TempDir("", "source")
		Expect(err).NotTo(HaveOccurred())

		Expect(ioutil.WriteFile(filepath.Join(sourceDir, "some-file"), []byte("some content"), 0644)).To(Succeed())

		Expect(os.MkdirAll(filepath.Join(sourceDir, "some-dir"), os.ModePerm)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(sourceDir, "some-dir", "other-file"), []byte("other content"), 0755)).To(Succeed())

		destDir, err = ioutil.TempDir("", "dest")
		Expect(err).NotTo(HaveOccurred())

		directoryDup = cargo.NewDirectoryDuplicator()
	})

	it.After(func() {
		Expect(os.RemoveAll(sourceDir)).To(Succeed())
		Expect(os.RemoveAll(destDir)).To(Succeed())
	})

	context("Duplicate", func() {
		it("duplicates the contents of a directory", func() {
			Expect(directoryDup.Duplicate(sourceDir, destDir)).To(Succeed())

			file, err := os.Open(filepath.Join(destDir, "some-file"))
			Expect(err).NotTo(HaveOccurred())

			content, err := ioutil.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("some content"))

			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode()).To(Equal(os.FileMode(0644)))

			Expect(file.Close()).To(Succeed())

			info, err = os.Stat(filepath.Join(destDir, "some-dir"))
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())

			file, err = os.Open(filepath.Join(destDir, "some-dir", "other-file"))
			Expect(err).NotTo(HaveOccurred())

			content, err = ioutil.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("other content"))

			info, err = file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode()).To(Equal(os.FileMode(0755)))

			Expect(file.Close()).To(Succeed())
		})
	})

	context("failure cases", func() {
		context("source dir does not exist", func() {
			it("fails", func() {
				err := directoryDup.Duplicate("does-not-exist", destDir)
				Expect(err).To(MatchError(ContainSubstring("source dir does not exist: ")))
			})
		})

		context("when source file has bad permissions", func() {
			it.Before(func() {
				Expect(os.Chmod(filepath.Join(sourceDir, "some-file"), 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(filepath.Join(sourceDir, "some-file"), 0644)).To(Succeed())
			})

			it("fails", func() {
				err := directoryDup.Duplicate(sourceDir, destDir)
				Expect(err).To(MatchError(ContainSubstring("opening source file failed:")))
			})
		})

		context("when destination directory bad permissions", func() {
			context("when creating dir", func() {
				it.Before(func() {
					Expect(os.Chmod(destDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(destDir, os.ModePerm)).To(Succeed())
				})

				it("fails", func() {
					err := directoryDup.Duplicate(sourceDir, destDir)
					Expect(err).To(MatchError(ContainSubstring("duplicate error creating dir:")))
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when creating file", func() {
				var dirPath string

				it.Before(func() {
					dirPath = filepath.Join(destDir, "some-dir")
					Expect(os.MkdirAll(dirPath, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(dirPath, os.ModePerm)).To(Succeed())
				})

				it("fails", func() {
					err := directoryDup.Duplicate(sourceDir, destDir)
					Expect(err).To(MatchError(ContainSubstring("duplicate error creating file:")))
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
}
