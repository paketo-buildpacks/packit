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

func testDeprecationWarning(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect       = NewWithT(t).Expect
		buildpackDir string

		deprecationWarning cargo.DeprecationWarning
	)
	it.Before(func() {
		var err error
		buildpackDir, err = ioutil.TempDir("", "buildpack")
		Expect(err).NotTo(HaveOccurred())

		deprecationWarning = cargo.NewDeprecationWarning()
	})

	it.After(func() {
		Expect(os.RemoveAll(buildpackDir)).To(Succeed())
	})

	context("WarnDepercatedFields", func() {
		context("when there are depercated fields", func() {
			it.Before(func() {
				err := ioutil.WriteFile(filepath.Join(buildpackDir, "buildpack.toml"), []byte(`api = "0.2"
[buildpack]
id = "some-buildpack-id"
name = "some-buildpack-name"
version = "some-buildpack-version"
homepage = "some-homepage-link"

[metadata]
include_files = ["some-include-file", "other-include-file"]
pre_package = "some-pre-package-script.sh"
`), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("fails if the buildpack.toml contains a depercated field", func() {
				err := deprecationWarning.WarnDeprecatedFields(filepath.Join(buildpackDir, "buildpack.toml"))
				Expect(err).To(MatchError("the include_files and pre_package feilds in the metadata section of the buildpack.toml have been changed to include-file and pre-package respectively: please update the buildpack.toml to refelct this change"))
			})
		})

		context("when there not are depercated fields", func() {
			it.Before(func() {
				err := ioutil.WriteFile(filepath.Join(buildpackDir, "buildpack.toml"), []byte(`api = "0.2"
[buildpack]
id = "some-buildpack-id"
name = "some-buildpack-name"
version = "some-buildpack-version"
homepage = "some-homepage-link"

[metadata]
include-files = ["some-include-file", "other-include-file"]
pre-package = "some-pre-package-script.sh"
`), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("fails if the buildpack.toml contains a depercated field", func() {
				err := deprecationWarning.WarnDeprecatedFields(filepath.Join(buildpackDir, "buildpack.toml"))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		context("fuilure cases", func() {
			it("returns an error", func() {
				err := deprecationWarning.WarnDeprecatedFields(filepath.Join(buildpackDir, "buildpack.toml"))
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})
		})
	})
}
