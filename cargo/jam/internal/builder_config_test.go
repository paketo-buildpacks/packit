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

func testBuilderConfig(t *testing.T, context spec.G, it spec.S) {
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
description = "Some description"

[[buildpacks]]
	uri = "docker://some-registry/some-repository/some-buildpack-id:0.0.10"
  version = "0.0.10"

[[buildpacks]]
	image = "some-registry/some-repository/other-buildpack-id:0.20.22"
  version = "0.20.22"

[lifecycle]
  version = "0.10.2"

[[order]]

  [[order.group]]
    id = "some-repository/other-buildpack-id"
    version = "0.20.22"

[[order]]

  [[order.group]]
    id = "some-repository/some-buildpack-id"
    version = "0.0.10"

[stack]
  id = "io.paketo.stacks.some-stack"
  build-image = "some-registry/somerepository/build:1.2.3-some-cnb"
  run-image = "some-registry/somerepository/run:some-cnb"
  run-image-mirrors = ["some-registry/some-repository/run:some-cnb"]
			`)
			Expect(err).NotTo(HaveOccurred())

			path = file.Name()
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		it("parses the builder.toml configuration", func() {
			config, err := internal.ParseBuilderConfig(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(internal.BuilderConfig{
				Description: "Some description",
				Buildpacks: []internal.BuilderConfigBuildpack{
					{
						URI:     "some-registry/some-repository/some-buildpack-id:0.0.10",
						Version: "0.0.10",
					},
					{
						URI:     "some-registry/some-repository/other-buildpack-id:0.20.22",
						Version: "0.20.22",
					},
				},
				Lifecycle: internal.BuilderConfigLifecycle{
					Version: "0.10.2",
				},
				Order: []internal.BuilderConfigOrder{
					{
						Group: []internal.BuilderConfigOrderGroup{
							{
								ID:      "some-repository/other-buildpack-id",
								Version: "0.20.22",
							},
						},
					},
					{
						Group: []internal.BuilderConfigOrderGroup{
							{
								ID:      "some-repository/some-buildpack-id",
								Version: "0.0.10",
							},
						},
					},
				},
				Stack: internal.BuilderConfigStack{
					ID:              "io.paketo.stacks.some-stack",
					BuildImage:      "some-registry/somerepository/build:1.2.3-some-cnb",
					RunImage:        "some-registry/somerepository/run:some-cnb",
					RunImageMirrors: []string{"some-registry/some-repository/run:some-cnb"},
				},
			}))
		})

		context("failure cases", func() {
			context("when the file cannot be opened", func() {
				it.Before(func() {
					Expect(os.Remove(path)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := internal.ParseBuilderConfig(path)
					Expect(err).To(MatchError(ContainSubstring("failed to open builder config file:")))
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})
		})

		context("when the file contents cannot be parsed", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte("%%%"), 0600)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := internal.ParseBuilderConfig(path)
				Expect(err).To(MatchError(ContainSubstring("failed to parse builder config:")))
				Expect(err).To(MatchError(ContainSubstring("keys cannot contain % character")))
			})
		})

		context("when a dependency uri is not valid", func() {
			it.Before(func() {
				Expect(ioutil.WriteFile(path, []byte(`
[[buildpacks]]
	uri = "docker://some-registry/some-repository/some-buildpack-id:0.0.10"
  version = "0.0.10"

[[buildpacks]]
	image = "some-registry/some-repository/other-buildpack-id:0.20.22"
  version = "0.20.22"

[[buildpacks]]
	uri = "%%%"
  version = "1.2.3"
					`), 0600)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := internal.ParseBuilderConfig(path)
				Expect(err).To(MatchError(ContainSubstring("failed to parse builder config:")))
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})
	})

	context("OverwritePackageConfig", func() {
		it.Before(func() {
			file, err := ioutil.TempFile("", "builder.toml")
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
			err := internal.OverwriteBuilderConfig(path, internal.BuilderConfig{
				Description: "Some description",
				Buildpacks: []internal.BuilderConfigBuildpack{
					{
						URI:     "some-registry/some-repository/some-buildpack-id:0.0.10",
						Version: "0.0.10",
					},
					{
						URI:     "some-registry/some-repository/other-buildpack-id:0.20.22",
						Version: "0.20.22",
					},
				},
				Lifecycle: internal.BuilderConfigLifecycle{
					Version: "0.10.2",
				},
				Order: []internal.BuilderConfigOrder{
					{
						Group: []internal.BuilderConfigOrderGroup{
							{
								ID:      "some-repository/other-buildpack-id",
								Version: "0.20.22",
							},
						},
					},
					{
						Group: []internal.BuilderConfigOrderGroup{
							{
								ID:      "some-repository/some-buildpack-id",
								Version: "0.0.10",
							},
						},
					},
				},
				Stack: internal.BuilderConfigStack{
					ID:              "io.paketo.stacks.some-stack",
					BuildImage:      "some-registry/somerepository/build:1.2.3-some-cnb",
					RunImage:        "some-registry/somerepository/run:some-cnb",
					RunImageMirrors: []string{"some-registry/some-repository/run:some-cnb"},
				},
			})
			Expect(err).NotTo(HaveOccurred())

			contents, err := ioutil.ReadFile(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(MatchTOML(`
description = "Some description"

[[buildpacks]]
	uri = "docker://some-registry/some-repository/some-buildpack-id:0.0.10"
  version = "0.0.10"

[[buildpacks]]
	uri = "docker://some-registry/some-repository/other-buildpack-id:0.20.22"
  version = "0.20.22"

[lifecycle]
  version = "0.10.2"

[[order]]

  [[order.group]]
    id = "some-repository/other-buildpack-id"
    version = "0.20.22"

[[order]]

  [[order.group]]
    id = "some-repository/some-buildpack-id"
    version = "0.0.10"

[stack]
  id = "io.paketo.stacks.some-stack"
  build-image = "some-registry/somerepository/build:1.2.3-some-cnb"
  run-image = "some-registry/somerepository/run:some-cnb"
  run-image-mirrors = ["some-registry/some-repository/run:some-cnb"]
				`))
		})

		context("failure cases", func() {
			context("when the file cannot be opened", func() {
				it.Before(func() {
					Expect(os.Remove(path)).To(Succeed())
				})

				it("returns an error", func() {
					err := internal.OverwriteBuilderConfig(path, internal.BuilderConfig{})
					Expect(err).To(MatchError(ContainSubstring("failed to open builder config file:")))
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})
		})
	})
}
