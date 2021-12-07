package fs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testExists(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		dirPath  string
		filePath string
	)

	context("Exists", func() {
		it.Before(func() {
			var err error
			dirPath, err = os.MkdirTemp("", "dir")
			Expect(err).NotTo(HaveOccurred())
			filePath = filepath.Join(dirPath, "some-file")
		})

		it.After(func() {
			Expect(os.RemoveAll(dirPath)).To(Succeed())
		})

		context("when the file exists", func() {
			it.Before(func() {
				Expect(os.WriteFile(filePath, []byte("hello file"), 0644)).To(Succeed())
			})

			it("returns true", func() {
				Expect(fs.Exists(filePath)).To(BeTrue())
			})
		})

		context("when the file DOES NOT exists", func() {
			it.Before(func() {
				Expect(os.RemoveAll(dirPath)).To(Succeed())
			})

			it("returns false", func() {
				Expect(fs.Exists(filePath)).To(BeFalse())
			})
		})

		context("when the file cannot be read", func() {
			it.Before(func() {
				Expect(os.WriteFile(filePath, nil, 0644)).To(Succeed())
				Expect(os.Chmod(dirPath, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(dirPath, os.ModePerm)).To(Succeed())
			})

			it("returns false and an error", func() {
				exists, err := fs.Exists(filePath)
				Expect(err.Error()).To(ContainSubstring("permission denied"))
				Expect(exists).To(BeFalse())
			})
		})
	})
}
