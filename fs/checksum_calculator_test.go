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

func testChecksumCalculator(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		calculator fs.ChecksumCalculator
		workingDir string
	)

	context("Sum", func() {
		it.Before(func() {
			var err error
			workingDir, err = ioutil.TempDir("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			calculator = fs.NewChecksumCalculator()
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		context("when given a single file", func() {
			var path string

			it.Before(func() {
				path = filepath.Join(workingDir, "some-file")
				Expect(ioutil.WriteFile(path, []byte{}, os.ModePerm)).To(Succeed())
			})

			it("generates the SHA256 checksum for that file", func() {
				sum, err := calculator.Sum(path)
				Expect(err).ToNot(HaveOccurred())
				Expect(sum).To(Equal("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"))
			})

			context("failure cases", func() {
				context("the file cannot be read", func() {
					it.Before(func() {
						Expect(os.Chmod(path, 0222)).To(Succeed())
					})

					it("returns an error", func() {
						_, err := calculator.Sum(path)
						Expect(err).To(MatchError(ContainSubstring("failed to calculate checksum")))
						Expect(err).To(MatchError(ContainSubstring("permission denied")))
					})
				})
			})
		})

		context("when given a directory", func() {
			var path string

			it.Before(func() {
				path = filepath.Join(workingDir, "some-dir")
				Expect(os.MkdirAll(path, os.ModePerm)).To(Succeed())

				Expect(ioutil.WriteFile(filepath.Join(path, "some-file"), []byte{}, os.ModePerm)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(path, "some-other-file"), []byte{}, os.ModePerm)).To(Succeed())
			})

			it("generates the SHA256 checksum for that directory", func() {
				sum, err := calculator.Sum(path)
				Expect(err).ToNot(HaveOccurred())
				Expect(sum).To(Equal("2dba5dbc339e7316aea2683faf839c1b7b1ee2313db792112588118df066aa35"))
			})

			context("failure cases", func() {
				context("when the directory cannot be read", func() {
					it.Before(func() {
						Expect(os.Chmod(path, 0222)).To(Succeed())
					})

					it.After(func() {
						Expect(os.Chmod(path, os.ModePerm)).To(Succeed())
					})

					it("returns an error", func() {
						_, err := calculator.Sum(path)
						Expect(err).To(MatchError(ContainSubstring("failed to calculate checksum")))
						Expect(err).To(MatchError(ContainSubstring("permission denied")))
					})
				})

				context("when a file in the directory cannot be read", func() {
					it.Before(func() {
						Expect(os.Chmod(filepath.Join(path, "some-file"), 0222)).To(Succeed())
					})

					it("returns an error", func() {
						_, err := calculator.Sum(path)
						Expect(err).To(MatchError(ContainSubstring("failed to calculate checksum")))
						Expect(err).To(MatchError(ContainSubstring("permission denied")))
					})
				})
			})
		})

		context("failure cases", func() {
			context("when the given path does not exist", func() {
				it("returns an error", func() {
					_, err := calculator.Sum("not a real path")
					Expect(err).To(MatchError(ContainSubstring("failed to calculate checksum")))
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})
		})
	})
}
