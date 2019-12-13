package cargo_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/packit/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDirectoryDuplicator(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect       = NewWithT(t).Expect
		tmpDir       string
		directoryDup cargo.DirectoryDuplicator
	)

	it.Before(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "dup-dest")
		Expect(err).NotTo(HaveOccurred())

		directoryDup = cargo.NewDirectoryDuplicator()
	})

	context("Duplicate", func() {
		it("duplicates the contents of a directory", func() {
			Expect(directoryDup.Duplicate(filepath.Join("jam", "testdata", "example-cnb"), tmpDir)).To(Succeed())

			var files []string
			err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !info.IsDir() {
					files = append(files, path)

					rel, err := filepath.Rel(tmpDir, path)
					Expect(err).NotTo(HaveOccurred())

					original, err := os.Stat(filepath.Join("jam", "testdata", "example-cnb", rel))
					Expect(err).NotTo(HaveOccurred())
					Expect(info.Mode()).To(Equal(original.Mode()))
				}

				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(HaveLen(4))
		})
	})
	context("Error cases", func() {

		context("source dir does not exist", func() {
			it("fails", func() {
				err := directoryDup.Duplicate(filepath.Join("jam", "testdata", "does-not-exits"), tmpDir)
				Expect(err).To(MatchError(ContainSubstring("source dir does not exist: ")))
			})
		})

		context("when source file has bad permissions", func() {
			it.Before(func() {
				Expect(os.Chmod(filepath.Join("jam", "testdata", "example-cnb", "buildpack.toml"), 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(filepath.Join("jam", "testdata", "example-cnb", "buildpack.toml"), 0644)).To(Succeed())
			})

			it("fails", func() {
				err := directoryDup.Duplicate(filepath.Join("jam", "testdata", "example-cnb"), tmpDir)
				Expect(err).To(MatchError(ContainSubstring("opening source file failed:")))
			})
		})

		context("when destination directory bad permissions", func() {
			context("when creating dir", func() {
				it.Before(func() {
					os.Chmod(tmpDir, 0000)
				})

				it.After(func() {
					os.Chmod(tmpDir, os.ModePerm)
				})
				it("fails", func() {
					err := directoryDup.Duplicate(filepath.Join("jam", "testdata", "example-cnb"), tmpDir)
					Expect(err).To(MatchError(ContainSubstring("duplicate error creating dir:")))
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when creating file", func() {
				var binPath string
				it.Before(func() {
					binPath = filepath.Join(tmpDir, "bin")
					os.MkdirAll(binPath, 0000)
				})

				it.After(func() {
					os.Chmod(binPath, os.ModePerm)
				})
				it("fails", func() {
					err := directoryDup.Duplicate(filepath.Join("jam", "testdata", "example-cnb"), tmpDir)
					Expect(err).To(MatchError(ContainSubstring("duplicate error creating file:")))
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
}
