package fs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testIsEmptyDir(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string
	)

	it.Before(func() {
		var err error
		path, err = os.MkdirTemp("", "dir")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("when the directory is empty", func() {
		it("returns true", func() {
			Expect(fs.IsEmptyDir(path)).To(BeTrue())
		})
	})

	context("when the directory is not empty", func() {
		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(path, "some-file"), []byte{}, 0644)).To(Succeed())
		})

		it("returns false", func() {
			Expect(fs.IsEmptyDir(path)).To(BeFalse())
		})
	})

	context("when the directory does not exist", func() {
		it.Before(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		it("returns false", func() {
			Expect(fs.IsEmptyDir(path)).To(BeFalse())
		})
	})

	context("when the directory cannot be read", func() {
		it.Before(func() {
			Expect(os.Chmod(path, 0000)).To(Succeed())
		})

		it.After(func() {
			Expect(os.Chmod(path, os.ModePerm)).To(Succeed())
		})

		it("returns false", func() {
			Expect(fs.IsEmptyDir(path)).To(BeFalse())
		})
	})
}
