package commands_test

import (
	"errors"
	"testing"

	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/cargo/jam/commands"
	"github.com/paketo-buildpacks/packit/cargo/jam/commands/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSummarize(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buildpackInspector *fakes.BuildpackInspector
		formatter          *fakes.Formatter

		command commands.Summarize
	)

	it.Before(func() {
		buildpackInspector = &fakes.BuildpackInspector{}
		buildpackInspector.DependenciesCall.Returns.ConfigSlice = []cargo.Config{
			{
				Buildpack: cargo.ConfigBuildpack{
					ID:      "some-buildpack",
					Version: "some-version",
				},
				Metadata: cargo.ConfigMetadata{
					Dependencies: []cargo.ConfigMetadataDependency{
						{
							ID:      "some-depency",
							Version: "some-version",
							Stacks:  []string{"some-stack"},
						},
					},
					DefaultVersions: map[string]string{
						"some-dependency": "some-version",
					},
				},
				Stacks: []cargo.ConfigStack{
					{ID: "some-stack"},
				},
			},
			{
				Buildpack: cargo.ConfigBuildpack{
					ID:      "other-buildpack",
					Version: "other-version",
				},
				Metadata: cargo.ConfigMetadata{
					Dependencies: []cargo.ConfigMetadataDependency{
						{
							ID:      "other-depency",
							Version: "other-version",
							Stacks:  []string{"other-stack"},
						},
					},
					DefaultVersions: map[string]string{
						"other-dependency": "other-version",
					},
				},
				Stacks: []cargo.ConfigStack{
					{ID: "other-stack"},
				},
			},
		}

		formatter = &fakes.Formatter{}

		command = commands.NewSummarize(buildpackInspector, formatter)
	})

	context("Execute", func() {
		it("prints a summary", func() {
			err := command.Execute([]string{
				"--buildpack", "buildpack.tgz",
				"--format", "markdown",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(buildpackInspector.DependenciesCall.Receives.Path).To(Equal("buildpack.tgz"))

			Expect(formatter.MarkdownCall.Receives.ConfigSlice).To(Equal([]cargo.Config{
				{
					Buildpack: cargo.ConfigBuildpack{
						ID:      "some-buildpack",
						Version: "some-version",
					},
					Metadata: cargo.ConfigMetadata{
						Dependencies: []cargo.ConfigMetadataDependency{
							{
								ID:      "some-depency",
								Version: "some-version",
								Stacks:  []string{"some-stack"},
							},
						},
						DefaultVersions: map[string]string{
							"some-dependency": "some-version",
						},
					},
					Stacks: []cargo.ConfigStack{
						{ID: "some-stack"},
					},
				},
				{
					Buildpack: cargo.ConfigBuildpack{
						ID:      "other-buildpack",
						Version: "other-version",
					},
					Metadata: cargo.ConfigMetadata{
						Dependencies: []cargo.ConfigMetadataDependency{
							{
								ID:      "other-depency",
								Version: "other-version",
								Stacks:  []string{"other-stack"},
							},
						},
						DefaultVersions: map[string]string{
							"other-dependency": "other-version",
						},
					},
					Stacks: []cargo.ConfigStack{
						{ID: "other-stack"},
					},
				},
			}))
		})

		context("when not given a --format flag", func() {
			it("prints a summary with the default of markdown", func() {
				err := command.Execute([]string{
					"--buildpack", "buildpack.tgz",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(buildpackInspector.DependenciesCall.Receives.Path).To(Equal("buildpack.tgz"))

				Expect(formatter.MarkdownCall.Receives.ConfigSlice).To(Equal([]cargo.Config{
					{
						Buildpack: cargo.ConfigBuildpack{
							ID:      "some-buildpack",
							Version: "some-version",
						},
						Metadata: cargo.ConfigMetadata{
							Dependencies: []cargo.ConfigMetadataDependency{
								{
									ID:      "some-depency",
									Version: "some-version",
									Stacks:  []string{"some-stack"},
								},
							},
							DefaultVersions: map[string]string{
								"some-dependency": "some-version",
							},
						},
						Stacks: []cargo.ConfigStack{
							{ID: "some-stack"},
						},
					},
					{
						Buildpack: cargo.ConfigBuildpack{
							ID:      "other-buildpack",
							Version: "other-version",
						},
						Metadata: cargo.ConfigMetadata{
							Dependencies: []cargo.ConfigMetadataDependency{
								{
									ID:      "other-depency",
									Version: "other-version",
									Stacks:  []string{"other-stack"},
								},
							},
							DefaultVersions: map[string]string{
								"other-dependency": "other-version",
							},
						},
						Stacks: []cargo.ConfigStack{
							{ID: "other-stack"},
						},
					},
				}))
			})
		})

		context("failure cases", func() {
			context("when given an unknown flag", func() {
				it("returns an error", func() {
					err := command.Execute([]string{"--unknown"})
					Expect(err).To(MatchError(ContainSubstring("flag provided but not defined: -unknown")))
				})
			})

			context("when the --buildpack flag is missing", func() {
				it("returns an error", func() {
					err := command.Execute(nil)
					Expect(err).To(MatchError("missing required flag --buildpack"))
				})
			})

			context("when buildpack inspector returns an error", func() {
				it.Before(func() {
					buildpackInspector.DependenciesCall.Returns.Error = errors.New("failed to get dependencies")
				})

				it("returns an error", func() {
					err := command.Execute([]string{
						"--buildpack", "buildpack.tgz",
					})
					Expect(err).To(MatchError("failed to inspect buildpack dependencies: failed to get dependencies"))
				})
			})

			context("when an unknown format is given to the --format flag", func() {
				it("returns an error", func() {
					err := command.Execute([]string{
						"--buildpack", "buildpack.tgz",
						"--format", "unknown",
					})
					Expect(err).To(MatchError(`unknown format "unknown", please choose from the following formats ("markdown")`))
				})
			})
		})
	})
}
