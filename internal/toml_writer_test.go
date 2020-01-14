package internal_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/packit/internal"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/packit/matchers"
	. "github.com/onsi/gomega"
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
		tmpDir, err = ioutil.TempDir("", "tomlWriter")
		Expect(err).NotTo(HaveOccurred())

		path = filepath.Join(tmpDir, "writer.toml")
	})

	it("writes the contents of a given object out to a .toml file", func() {
		err := tomlWriter.Write(path, map[string]string{
			"some-field":  "some-value",
			"other-field": "other-value",
		})
		Expect(err).NotTo(HaveOccurred())

		tomlFileContents, err := ioutil.ReadFile(path)
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
				err := tomlWriter.Write(path, map[int]int{1: 100})
				Expect(err).To(MatchError(ContainSubstring("cannot encode a map with non-string key type")))
			})
		})
	})
}
