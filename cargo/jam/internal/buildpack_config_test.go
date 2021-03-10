package internal_test

import (
	"os"
	"testing"

	"github.com/paketo-buildpacks/packit/cargo/jam/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/packit/matchers"
)

func testBuildpackConfig(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string
	)

	context("ParseBuildpackConfig", func() {
		it.Before(func() {
			file, err := os.CreateTemp("", "buildpack.toml")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			_, err = file.WriteString(`
				api = "0.2"

				[buildpack]
					id = "some-composite-buildpack"
					name = "Some Composite Buildpack"
					version = "some-composite-buildpack-version"

				[metadata]
					include-files = ["buildpack.toml"]

				[[order]]
					[[order.group]]
						id = "some-repository/some-buildpack-id"
						version = "0.20.1"

					[[order.group]]
						id = "some-repository/last-buildpack-id"
						version = "0.2.0"

				[[order]]
					[[order.group]]
						id = "some-repository/other-buildpack-id"
						version = "0.1.0"
						optional = true
			`)
			Expect(err).NotTo(HaveOccurred())

			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		it("parses the buildpack.toml configuration", func() {
			config, err := internal.ParseBuildpackConfig(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(internal.BuildpackConfig{
				API: "0.2",
				Buildpack: map[string]interface{}{
					"id":      "some-composite-buildpack",
					"name":    "Some Composite Buildpack",
					"version": "some-composite-buildpack-version",
				},
				Metadata: map[string]interface{}{
					"include-files": []interface{}{"buildpack.toml"},
				},
				Order: []internal.BuildpackConfigOrder{
					{
						Group: []internal.BuildpackConfigOrderGroup{
							{
								ID:      "some-repository/some-buildpack-id",
								Version: "0.20.1",
							},
							{
								ID:      "some-repository/last-buildpack-id",
								Version: "0.2.0",
							},
						},
					},
					{
						Group: []internal.BuildpackConfigOrderGroup{
							{
								ID:       "some-repository/other-buildpack-id",
								Version:  "0.1.0",
								Optional: true,
							},
						},
					},
				},
			}))
		})

		context("failure cases", func() {
			context("when the file cannot be opened", func() {
				it.Before(func() {
					Expect(os.Remove(path)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := internal.ParseBuildpackConfig(path)
					Expect(err).To(MatchError(ContainSubstring("failed to open buildpack config file:")))
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			context("when the file contents cannot be parsed", func() {
				it.Before(func() {
					Expect(os.WriteFile(path, []byte("%%%"), 0600)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := internal.ParseBuildpackConfig(path)
					Expect(err).To(MatchError(ContainSubstring("failed to parse buildpack config:")))
					Expect(err).To(MatchError(ContainSubstring("keys cannot contain % character")))
				})
			})
		})
	})

	context("OverwriteBuildpackConfig", func() {
		it.Before(func() {
			file, err := os.CreateTemp("", "buildpack.toml")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			_, err = file.WriteString(`previous contents of the file`)
			Expect(err).NotTo(HaveOccurred())

			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		it("overwrites the buildpack.toml configuration", func() {
			err := internal.OverwriteBuildpackConfig(path, internal.BuildpackConfig{
				API: "0.2",
				Buildpack: map[string]interface{}{
					"id":      "some-composite-buildpack",
					"name":    "Some Composite Buildpack",
					"version": "some-composite-buildpack-version",
				},
				Metadata: map[string]interface{}{
					"include-files": []interface{}{"buildpack.toml"},
				},
				Order: []internal.BuildpackConfigOrder{
					{
						Group: []internal.BuildpackConfigOrderGroup{
							{
								ID:      "some-repository/some-buildpack-id",
								Version: "0.20.1",
							},
							{
								ID:      "some-repository/last-buildpack-id",
								Version: "0.2.0",
							},
						},
					},
					{
						Group: []internal.BuildpackConfigOrderGroup{
							{
								ID:       "some-repository/other-buildpack-id",
								Version:  "0.1.0",
								Optional: true,
							},
						},
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())

			contents, err := os.ReadFile(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(MatchTOML(`
				api = "0.2"

				[buildpack]
					id = "some-composite-buildpack"
					name = "Some Composite Buildpack"
					version = "some-composite-buildpack-version"

				[metadata]
					include-files = ["buildpack.toml"]

				[[order]]
					[[order.group]]
						id = "some-repository/some-buildpack-id"
						version = "0.20.1"

					[[order.group]]
						id = "some-repository/last-buildpack-id"
						version = "0.2.0"

				[[order]]
					[[order.group]]
						id = "some-repository/other-buildpack-id"
						version = "0.1.0"
						optional = true
			`))
		})

		context("failure cases", func() {
			context("when the file cannot be opened", func() {
				it.Before(func() {
					Expect(os.Remove(path)).To(Succeed())
				})

				it("returns an error", func() {
					err := internal.OverwriteBuildpackConfig(path, internal.BuildpackConfig{})
					Expect(err).To(MatchError(ContainSubstring("failed to open buildpack config file:")))
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			context("when the config cannot be written", func() {
				it("returns an error", func() {
					err := internal.OverwriteBuildpackConfig(path, internal.BuildpackConfig{
						API: func() {},
					})
					Expect(err).To(MatchError(ContainSubstring("failed to write buildpack config:")))
					Expect(err).To(MatchError(ContainSubstring("Marshal can't handle func()(func)")))
				})
			})
		})
	})
}
