package est_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/packit/est"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testFileExists(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect   = NewWithT(t).Expect
		tempDir  string
		tempFile string
	)

	it.Before(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "some-dir")
		Expect(err).ToNot(HaveOccurred())

		tempFile = filepath.Join(tempDir, "some-file")
	})

	it.After(func() {
		Expect(os.RemoveAll(tempDir)).To(Succeed())
	})

	context("FileExists", func() {
		context("when the file exists", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(tempFile, []byte(`some contents`), 0644)).To(Succeed())
			})
			it("return true", func() {
				exists, err := est.FileExists(tempFile)
				Expect(err).NotTo(HaveOccurred())

				Expect(exists).To(BeTrue())
			})
		})

		context("file does not exist", func() {
			it("returns false", func() {
				exists, err := est.FileExists(tempFile)
				Expect(err).NotTo(HaveOccurred())

				Expect(exists).To(BeFalse())
			})
		})

		context("failure cases", func() {

			context("it can not read, write, or execute inside dir where file is", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(tempFile, []byte(`some contents`), 0644)).To(Succeed())
					Expect(os.Chmod(tempDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(tempDir, os.ModePerm)).To(Succeed())
				})

				it("returns false", func() {
					_, err := est.FileExists(tempFile)
					Expect(err).To(MatchError(ContainSubstring("could not stat file:")))
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
}
