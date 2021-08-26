package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/packit/matchers"
)

func testTOMLWriter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		tmpDir     string
		path       string
		tomlWriter internal.TOMLWriter
	)
	it.Before(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "tomlWriter")
		Expect(err).NotTo(HaveOccurred())

		path = filepath.Join(tmpDir, "writer.toml")
	})

	it("writes the contents of a given object out to a .toml file", func() {
		err := tomlWriter.Write(path, map[string]string{
			"some-field":  "some-value",
			"other-field": "other-value",
		})
		Expect(err).NotTo(HaveOccurred())

		tomlFileContents, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(tomlFileContents)).To(MatchTOML(`
some-field = "some-value"
other-field = "other-value"`))
	})

	context("failure cases", func() {
		context("the .toml file cannot be created", func() {
			it.Before(func() {
				Expect(os.RemoveAll(tmpDir)).To(Succeed())
			})

			it("returns an error", func() {
				err := tomlWriter.Write(path, map[string]string{
					"some-field":  "some-value",
					"other-field": "other-value",
				})
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})
		})

		context("the TOML data is invalid", func() {

			it("returns an error", func() {
				err := tomlWriter.Write(path, 0)
				Expect(err).To(MatchError(ContainSubstring("Only a struct or map can be marshaled to TOML")))
			})
		})
	})
}
