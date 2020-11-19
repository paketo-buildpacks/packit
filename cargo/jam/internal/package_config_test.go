package internal_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/paketo-buildpacks/packit/cargo/jam/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/packit/matchers"
)

func testPackageConfig(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string
	)

	context("ParsePackageConfig", func() {
		it.Before(func() {
			file, err := ioutil.TempFile("", "package.toml")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			_, err = file.WriteString(`
				[buildpack]
				uri = "build/buildpack.tgz"

				[[dependencies]]
				uri = "docker://some-registry/some-repository/last-buildpack-id:0.2.0"

				[[dependencies]]
				image = "some-registry/some-repository/some-buildpack-id:0.20.1"

				[[dependencies]]
				uri = "docker://some-registry/some-repository/other-buildpack-id:0.1.0"
			`)
			Expect(err).NotTo(HaveOccurred())

			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		it("parses the package.toml configuration", func() {
			config, err := internal.ParsePackageConfig(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(internal.PackageConfig{
				Buildpack: map[string]interface{}{
					"uri": "build/buildpack.tgz",
				},
				Dependencies: []internal.PackageConfigDependency{
					{URI: "some-registry/some-repository/last-buildpack-id:0.2.0"},
					{URI: "some-registry/some-repository/some-buildpack-id:0.20.1"},
					{URI: "some-registry/some-repository/other-buildpack-id:0.1.0"},
				},
			}))
		})

		context("failure cases", func() {
			context("when the file cannot be opened", func() {
				it.Before(func() {
					Expect(os.Remove(path)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := internal.ParsePackageConfig(path)
					Expect(err).To(MatchError(ContainSubstring("failed to open package config file:")))
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			context("when the file contents cannot be parsed", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(path, []byte("%%%"), 0600)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := internal.ParsePackageConfig(path)
					Expect(err).To(MatchError(ContainSubstring("failed to parse package config:")))
					Expect(err).To(MatchError(ContainSubstring("keys cannot contain % character")))
				})
			})

			context("when a dependency uri is not valid", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(path, []byte(`
						[buildpack]
						uri = "build/buildpack.tgz"

						[[dependencies]]
						uri = "docker://some-registry/some-repository/last-buildpack-id:0.2.0"

						[[dependencies]]
						image = "some-registry/some-repository/some-buildpack-id:0.20.1"

						[[dependencies]]
						uri = "%%%"
					`), 0600)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := internal.ParsePackageConfig(path)
					Expect(err).To(MatchError(ContainSubstring("failed to parse package config:")))
					Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
				})
			})
		})
	})

	context("OverwritePackageConfig", func() {
		it.Before(func() {
			file, err := ioutil.TempFile("", "package.toml")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			_, err = file.WriteString(`previous contents of the file`)
			Expect(err).NotTo(HaveOccurred())

			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		it("overwrites the package.toml configuration", func() {
			err := internal.OverwritePackageConfig(path, internal.PackageConfig{
				Buildpack: map[string]interface{}{
					"uri": "build/buildpack.tgz",
				},
				Dependencies: []internal.PackageConfigDependency{
					{URI: "some-registry/some-repository/last-buildpack-id:0.2.0"},
					{URI: "some-registry/some-repository/some-buildpack-id:0.20.1"},
					{URI: "some-registry/some-repository/other-buildpack-id:0.1.0"},
				},
			})
			Expect(err).NotTo(HaveOccurred())

			contents, err := ioutil.ReadFile(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(MatchTOML(`
				[buildpack]
				uri = "build/buildpack.tgz"

				[[dependencies]]
				uri = "docker://some-registry/some-repository/last-buildpack-id:0.2.0"

				[[dependencies]]
				uri = "docker://some-registry/some-repository/some-buildpack-id:0.20.1"

				[[dependencies]]
				uri = "docker://some-registry/some-repository/other-buildpack-id:0.1.0"
			`))
		})

		context("failure cases", func() {
			context("when the file cannot be opened", func() {
				it.Before(func() {
					Expect(os.Remove(path)).To(Succeed())
				})

				it("returns an error", func() {
					err := internal.OverwritePackageConfig(path, internal.PackageConfig{})
					Expect(err).To(MatchError(ContainSubstring("failed to open package config file:")))
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			context("when the config cannot be written", func() {
				it("returns an error", func() {
					err := internal.OverwritePackageConfig(path, internal.PackageConfig{
						Buildpack: func() {},
					})
					Expect(err).To(MatchError(ContainSubstring("failed to write package config:")))
					Expect(err).To(MatchError(ContainSubstring("Marshal can't handle func()(func)")))
				})
			})
		})
	})
}
