package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testEnvironmentWriter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		tmpDir string
		writer internal.EnvironmentWriter
	)

	it.Before(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "env-vars")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.RemoveAll(tmpDir)).To(Succeed())

		writer = internal.NewEnvironmentWriter()
	})

	it.After(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	it("writes the given environment to a directory", func() {
		err := writer.Write(tmpDir, map[string]string{
			"some-name":  "some-content",
			"other-name": "other-content",
		})
		Expect(err).NotTo(HaveOccurred())

		content, err := os.ReadFile(filepath.Join(tmpDir, "some-name"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("some-content"))

		content, err = os.ReadFile(filepath.Join(tmpDir, "other-name"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("other-content"))
	})

	it("writes does not create a directory of the env map is empty", func() {
		err := writer.Write(tmpDir, map[string]string{})
		Expect(err).NotTo(HaveOccurred())

		Expect(tmpDir).NotTo(BeAnExistingFile())
	})

	context("failure cases", func() {
		context("when the directory cannot be created", func() {
			it.Before(func() {
				Expect(os.MkdirAll(tmpDir, 0000)).To(Succeed())
			})

			it("returns an error", func() {
				err := writer.Write(filepath.Join(tmpDir, "sub-dir"), map[string]string{
					"some-name":  "some-content",
					"other-name": "other-content",
				})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the env file cannot be created", func() {
			it.Before(func() {
				Expect(os.MkdirAll(tmpDir, 0000)).To(Succeed())
			})

			it("returns an error", func() {
				err := writer.Write(tmpDir, map[string]string{
					"some-name":  "some-content",
					"other-name": "other-content",
				})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})
}
