package packit_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/packit/matchers"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir  string
		platformDir string
		tmpDir      string
		layersDir   string
		planPath    string
		cnbDir      string
		envCnbDir   string
		binaryPath  string
		exitHandler *fakes.ExitHandler
	)

	it.Before(func() {
		var err error
		workingDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		tmpDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		tmpDir, err = filepath.EvalSymlinks(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		Expect(os.Chdir(tmpDir)).To(Succeed())

		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		platformDir, err = os.MkdirTemp("", "platform")
		Expect(err).NotTo(HaveOccurred())

		file, err := os.CreateTemp("", "plan.toml")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		_, err = file.WriteString(`
[[entries]]
  name = "some-entry"

[entries.metadata]
  version = "some-version"
  some-key = "some-value"
`)
		Expect(err).NotTo(HaveOccurred())

		planPath = file.Name()

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		envCnbDir, err = os.MkdirTemp("", "envCnb")
		Expect(err).NotTo(HaveOccurred())

		bpTOML := []byte(`
api = "0.7"
[buildpack]
  id = "some-id"
  name = "some-name"
  version = "some-version"
	homepage = "some-homepage"
	description = "some-description"
	keywords = ["some-keyword"]
  clear-env = false

	[[buildpack.licenses]]
		type = "some-license-type"
		uri = "some-license-uri"
`)
		Expect(os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(envCnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())

		binaryPath = filepath.Join(cnbDir, "bin", "build")

		Expect(os.Setenv("CNB_STACK_ID", "some-stack")).To(Succeed())

		exitHandler = &fakes.ExitHandler{}
	})

	it.After(func() {
		Expect(os.Unsetenv("CNB_STACK_ID")).To(Succeed())

		Expect(os.Chdir(workingDir)).To(Succeed())
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(platformDir)).To(Succeed())
	})

	it("provides the build context to the given BuildFunc", func() {
		var context packit.BuildContext

		packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
			context = ctx

			return packit.BuildResult{}, nil
		}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

		Expect(context).To(Equal(packit.BuildContext{
			CNBPath: cnbDir,
			Stack:   "some-stack",
			Platform: packit.Platform{
				Path: platformDir,
			},
			WorkingDir: tmpDir,
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "some-entry",
						Metadata: map[string]interface{}{
							"version":  "some-version",
							"some-key": "some-value",
						},
					},
				},
			},
			Layers: packit.Layers{
				Path: layersDir,
			},
			BuildpackInfo: packit.BuildpackInfo{
				ID:          "some-id",
				Name:        "some-name",
				Version:     "some-version",
				Homepage:    "some-homepage",
				Description: "some-description",
				Keywords:    []string{"some-keyword"},
				Licenses: []packit.BuildpackInfoLicense{
					{
						Type: "some-license-type",
						URI:  "some-license-uri",
					},
				},
			},
		}))
	})

	context("when there are updates to the build plan", func() {
		context("when the api version is less than 0.5", func() {
			it.Before(func() {
				bpTOML := []byte(`
api = "0.4"
[buildpack]
  id = "some-id"
  name = "some-name"
  version = "some-version"
  clear-env = false
`)
				Expect(os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(envCnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())

			})

			it("updates the buildpack plan.toml with any changes", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					ctx.Plan.Entries[0].Metadata["other-key"] = "other-value"

					return packit.BuildResult{
						Plan: ctx.Plan,
					}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

				contents, err := os.ReadFile(planPath)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(contents)).To(MatchTOML(`
[[entries]]
  name = "some-entry"

[entries.metadata]
  version = "some-version"
  some-key = "some-value"
  other-key = "other-value"
`))
			})
		})

		context("when the api version is greater or equal to 0.5", func() {
			it("throws an error", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{
						Plan: ctx.Plan,
					}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("buildpack plan is read only")))
			})
		})
	})

	it("persists layer metadata", func() {
		packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
			layerPath := filepath.Join(ctx.Layers.Path, "some-layer")
			Expect(os.MkdirAll(layerPath, os.ModePerm)).To(Succeed())

			return packit.BuildResult{
				Layers: []packit.Layer{
					packit.Layer{
						Path:   layerPath,
						Name:   "some-layer",
						Build:  true,
						Launch: true,
						Cache:  true,
						Metadata: map[string]interface{}{
							"some-key": "some-value",
						},
					},
				},
			}, nil
		}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

		contents, err := os.ReadFile(filepath.Join(layersDir, "some-layer.toml"))
		Expect(err).NotTo(HaveOccurred())

		Expect(string(contents)).To(MatchTOML(`
[types]
  launch = true
  build = true
  cache = true

[metadata]
  some-key = "some-value"
`))
	})

	context("when the buildpack api version is less than 0.6", func() {
		it.Before(func() {
			bpTOML := []byte(`
api = "0.5"
[buildpack]
  id = "some-id"
  name = "some-name"
  version = "some-version"
`)
			Expect(os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())
		})

		it("persists layer metadata", func() {
			packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
				layerPath := filepath.Join(ctx.Layers.Path, "some-layer")
				Expect(os.MkdirAll(layerPath, os.ModePerm)).To(Succeed())

				return packit.BuildResult{
					Layers: []packit.Layer{
						packit.Layer{
							Path:   layerPath,
							Name:   "some-layer",
							Build:  true,
							Launch: true,
							Cache:  true,
							Metadata: map[string]interface{}{
								"some-key": "some-value",
							},
						},
					},
				}, nil
			}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

			contents, err := os.ReadFile(filepath.Join(layersDir, "some-layer.toml"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(contents)).To(MatchTOML(`
launch = true
build = true
cache = true

[metadata]
  some-key = "some-value"
`))
		})
	})

	context("when there are sbom entries layer metadata", func() {
		it("writes them to their specified locations", func() {
			packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
				layerPath := filepath.Join(ctx.Layers.Path, "some-layer")
				Expect(os.MkdirAll(layerPath, os.ModePerm)).To(Succeed())

				return packit.BuildResult{
					Layers: []packit.Layer{
						packit.Layer{
							Path: layerPath,
							Name: "some-layer",
							SBOM: packit.SBOMEntries{
								"some.json": strings.NewReader(`{"some-key": "some-value"}`),
								"other.yml": strings.NewReader(`other-key: other-value`),
							},
						},
					},
				}, nil
			}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

			contents, err := os.ReadFile(filepath.Join(layersDir, "some-layer.sbom.some.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(MatchJSON(`{"some-key": "some-value"}`))

			contents, err = os.ReadFile(filepath.Join(layersDir, "some-layer.sbom.other.yml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(MatchYAML(`other-key: other-value`))
		})

		context("when the api version is less than 0.7", func() {
			it.Before(func() {
				bpTOML := []byte(`
api = "0.6"
[buildpack]
  id = "some-id"
  name = "some-name"
  version = "some-version"
  clear-env = false
`)
				Expect(os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(envCnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())
			})

			it("throws an error", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					layerPath := filepath.Join(ctx.Layers.Path, "some-layer")
					Expect(os.MkdirAll(layerPath, os.ModePerm)).To(Succeed())

					return packit.BuildResult{
						Layers: []packit.Layer{
							packit.Layer{
								Path: layerPath,
								Name: "some-layer",
								SBOM: packit.SBOMEntries{
									"some.json": strings.NewReader(`{"some-key": "some-value"}`),
									"other.yml": strings.NewReader(`other-key: other-value`),
								},
							},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))
				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("some-layer.sbom.* output is only supported with Buildpack API v0.7 or higher")))
			})
		})
	})

	context("when there are existing layer.toml files", func() {
		context("when the layer.toml's will not be re-written", func() {
			var obsoleteLayerPath string

			it.Before(func() {
				obsoleteLayerPath = filepath.Join(layersDir, "obsolete-layer")
				Expect(os.MkdirAll(obsoleteLayerPath, os.ModePerm)).To(Succeed())
				Expect(os.WriteFile(obsoleteLayerPath+".toml", []byte{}, 0600)).To(Succeed())

				Expect(os.WriteFile(filepath.Join(layersDir, "launch.toml"), []byte{}, 0600)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(layersDir, "store.toml"), []byte{}, 0600)).To(Succeed())
			})

			context("when the buildpack api version is less than 0.6", func() {
				it.Before(func() {
					bpTOML := []byte(`
api = "0.5"
[buildpack]
  id = "some-id"
  name = "some-name"
  version = "some-version"
`)
					Expect(os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())
				})

				it("removes them", func() {
					packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
						return packit.BuildResult{
							Layers: []packit.Layer{},
						}, nil
					}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))
					Expect(obsoleteLayerPath).NotTo(BeARegularFile())
					Expect(obsoleteLayerPath + ".toml").NotTo(BeARegularFile())

					Expect(filepath.Join(layersDir, "launch.toml")).To(BeARegularFile())
					Expect(filepath.Join(layersDir, "store.toml")).To(BeARegularFile())
				})

				context("failures", func() {
					context("when getting the layer toml list", func() {
						var unremovableTOMLPath string

						it.Before(func() {
							unremovableTOMLPath = filepath.Join(layersDir, "unremovable.toml")
							Expect(os.MkdirAll(filepath.Join(layersDir, "unremovable"), os.ModePerm)).To(Succeed())
							Expect(os.WriteFile(unremovableTOMLPath, []byte{}, os.ModePerm)).To(Succeed())
							Expect(os.Chmod(layersDir, 0666)).To(Succeed())
						})

						it.After(func() {
							Expect(os.Chmod(layersDir, os.ModePerm)).To(Succeed())
						})

						it("returns an error", func() {
							packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
								return packit.BuildResult{
									Layers: []packit.Layer{},
								}, nil
							}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))
							Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("failed to remove layer toml:")))
						})
					})
				})
			})

			context("when the buildpack api version is greater than or equal to 0.6", func() {
				it.Before(func() {
					bpTOML := []byte(`
api = "0.6"
[buildpack]
  id = "some-id"
  name = "some-name"
  version = "some-version"
`)
					Expect(os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())
				})

				it("leaves them in place", func() {
					packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
						return packit.BuildResult{
							Layers: []packit.Layer{},
						}, nil
					}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

					Expect(obsoleteLayerPath).To(BeADirectory())
					Expect(obsoleteLayerPath + ".toml").To(BeARegularFile())

					Expect(filepath.Join(layersDir, "launch.toml")).To(BeARegularFile())
					Expect(filepath.Join(layersDir, "store.toml")).To(BeARegularFile())
				})
			})
		})
	})

	context("when the CNB_BUILDPACK_DIR environment variable is set", func() {
		it.Before(func() {
			os.Setenv("CNB_BUILDPACK_DIR", envCnbDir)
		})

		it.After(func() {
			os.Unsetenv("CNB_BUILDPACK_DIR")
		})

		it("sets the correct value for CNBdir in the Build context", func() {
			var context packit.BuildContext

			packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
				context = ctx

				return packit.BuildResult{}, nil
			}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

			Expect(context).To(Equal(packit.BuildContext{
				CNBPath: envCnbDir,
				Platform: packit.Platform{
					Path: platformDir,
				},
				Stack:      "some-stack",
				WorkingDir: tmpDir,
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "some-entry",
							Metadata: map[string]interface{}{
								"version":  "some-version",
								"some-key": "some-value",
							},
						},
					},
				},
				Layers: packit.Layers{
					Path: layersDir,
				},
				BuildpackInfo: packit.BuildpackInfo{
					ID:          "some-id",
					Name:        "some-name",
					Version:     "some-version",
					Homepage:    "some-homepage",
					Description: "some-description",
					Keywords:    []string{"some-keyword"},
					Licenses: []packit.BuildpackInfoLicense{
						{
							Type: "some-license-type",
							URI:  "some-license-uri",
						},
					},
				},
			}))
		})
	})

	context("when there are sbom entries in the build metadata", func() {
		it("writes them to their specified locations", func() {
			packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
				return packit.BuildResult{
					Build: packit.BuildMetadata{
						SBOM: packit.SBOMEntries{
							"some.json": strings.NewReader(`{"some-key": "some-value"}`),
							"other.yml": strings.NewReader(`other-key: other-value`),
						},
					},
				}, nil
			}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

			contents, err := os.ReadFile(filepath.Join(layersDir, "build.sbom.some.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(MatchJSON(`{"some-key": "some-value"}`))

			contents, err = os.ReadFile(filepath.Join(layersDir, "build.sbom.other.yml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(MatchYAML(`other-key: other-value`))
		})

		context("when the api version is less than 0.7", func() {
			it.Before(func() {
				bpTOML := []byte(`
api = "0.6"
[buildpack]
  id = "some-id"
  name = "some-name"
  version = "some-version"
  clear-env = false
`)
				Expect(os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(envCnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())
			})

			it("throws an error", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{
						Build: packit.BuildMetadata{
							SBOM: packit.SBOMEntries{
								"some.json": strings.NewReader(`{"some-key": "some-value"}`),
								"other.yml": strings.NewReader(`other-key: other-value`),
							},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))
				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("build.sbom.* output is only supported with Buildpack API v0.7 or higher")))
			})
		})
	})

	context("when there are bom entries in the build metadata", func() {
		it("persists a build.toml", func() {
			algorithm512, err := packit.GetBOMChecksumAlgorithm("sha512")
			Expect(err).ToNot(HaveOccurred())

			packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
				return packit.BuildResult{
					Build: packit.BuildMetadata{
						BOM: []packit.BOMEntry{
							{
								Name: "example",
							},
							{
								Name: "another-example",
								Metadata: packit.BOMMetadata{
									Version: "0.5",
									Checksum: packit.BOMChecksum{
										Algorithm: packit.SHA256,
										Hash:      "12345",
									},
									Source: packit.BOMSource{
										Checksum: packit.BOMChecksum{
											Algorithm: algorithm512,
											Hash:      "some-source-sha",
										},
									},
								},
							},
						},
					},
				}, nil
			}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

			contents, err := os.ReadFile(filepath.Join(layersDir, "build.toml"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(contents)).To(MatchTOML(`
				[[bom]]
					name = "example"
				[[bom]]
					name = "another-example"
				[bom.metadata]
					version = "0.5"
				[bom.metadata.checksum]
					algorithm = "SHA-256"
					hash = "12345"
				[bom.metadata.source]
					[bom.metadata.source.checksum]
						algorithm = "SHA-512"
						hash = "some-source-sha"
			`))
		})

		context("when the api version is less than 0.5", func() {
			it.Before(func() {
				bpTOML := []byte(`
api = "0.4"
[buildpack]
  id = "some-id"
  name = "some-name"
  version = "some-version"
  clear-env = false
`)
				Expect(os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(envCnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())
			})

			it("throws an error", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{
						Build: packit.BuildMetadata{
							BOM: []packit.BOMEntry{
								{
									Name: "example",
								},
								{
									Name: "another-example",
									Metadata: packit.BOMMetadata{
										Version: "0.5",
									},
								},
							},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))
				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("build.toml is only supported with Buildpack API v0.5 or higher")))
			})
		})
	})

	context("when there are unmet entries in the build metadata", func() {
		it("persists a build.toml", func() {
			packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
				return packit.BuildResult{
					Build: packit.BuildMetadata{
						Unmet: []packit.UnmetEntry{
							{
								Name: "example",
							},
							{
								Name: "another-example",
							},
						},
					},
				}, nil
			}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

			contents, err := os.ReadFile(filepath.Join(layersDir, "build.toml"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(contents)).To(MatchTOML(`
				[[unmet]]
					name = "example"
				[[unmet]]
					name = "another-example"
			`))
		})
		context("when the api version is less than 0.5", func() {
			it.Before(func() {
				bpTOML := []byte(`
api = "0.4"
[buildpack]
  id = "some-id"
  name = "some-name"
  version = "some-version"
  clear-env = false
`)
				Expect(os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(envCnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())

			})

			it("throws an error", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{
						Build: packit.BuildMetadata{
							Unmet: []packit.UnmetEntry{
								{
									Name: "example",
								},
								{
									Name: "another-example",
								},
							},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))
				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("build.toml is only supported with Buildpack API v0.5 or higher")))

			})
		})
	})

	context("when there are sbom entries in the launch metadata", func() {
		it("writes them to their specified locations", func() {
			packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
				return packit.BuildResult{
					Launch: packit.LaunchMetadata{
						SBOM: packit.SBOMEntries{
							"some.json": strings.NewReader(`{"some-key": "some-value"}`),
							"other.yml": strings.NewReader(`other-key: other-value`),
						},
					},
				}, nil
			}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

			contents, err := os.ReadFile(filepath.Join(layersDir, "launch.sbom.some.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(MatchJSON(`{"some-key": "some-value"}`))

			contents, err = os.ReadFile(filepath.Join(layersDir, "launch.sbom.other.yml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(MatchYAML(`other-key: other-value`))
		})

		context("when the api version is less than 0.7", func() {
			it.Before(func() {
				bpTOML := []byte(`
api = "0.6"
[buildpack]
  id = "some-id"
  name = "some-name"
  version = "some-version"
  clear-env = false
`)
				Expect(os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(envCnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())
			})

			it("throws an error", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{
						Launch: packit.LaunchMetadata{
							SBOM: packit.SBOMEntries{
								"some.json": strings.NewReader(`{"some-key": "some-value"}`),
								"other.yml": strings.NewReader(`other-key: other-value`),
							},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))
				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("launch.sbom.* output is only supported with Buildpack API v0.7 or higher")))
			})
		})
	})

	context("when there are bom entries in the launch metadata", func() {
		it("persists a launch.toml", func() {
			packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
				return packit.BuildResult{
					Launch: packit.LaunchMetadata{
						BOM: []packit.BOMEntry{
							{
								Name: "example",
							},
							{
								Name: "another-example",
								Metadata: packit.BOMMetadata{
									Version: "0.5",
								},
							},
						},
					},
				}, nil
			}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

			contents, err := os.ReadFile(filepath.Join(layersDir, "launch.toml"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(contents)).To(MatchTOML(`
				[[bom]]
					name = "example"
				[[bom]]
					name = "another-example"
				[bom.metadata]
					version = "0.5"
			`))
		})
	})

	context("when there are processes in the result", func() {
		it("persists a launch.toml", func() {
			packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
				return packit.BuildResult{
					Launch: packit.LaunchMetadata{
						Processes: []packit.Process{
							{
								Type:    "some-type",
								Command: "some-command",
								Args:    []string{"some-arg"},
								Direct:  true,
							},
						},
					},
				}, nil
			}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

			contents, err := os.ReadFile(filepath.Join(layersDir, "launch.toml"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(contents)).To(MatchTOML(`
				[[processes]]
					type = "some-type"
					command = "some-command"
					args = ["some-arg"]
					direct = true
			`))
		})

		context("when the process is the default", func() {
			it("persists a launch.toml", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{
						Launch: packit.LaunchMetadata{
							Processes: []packit.Process{
								{
									Type:    "some-type",
									Command: "some-command",
									Args:    []string{"some-arg"},
									Direct:  true,
									Default: true,
								},
							},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

				contents, err := os.ReadFile(filepath.Join(layersDir, "launch.toml"))
				Expect(err).NotTo(HaveOccurred())

				Expect(string(contents)).To(MatchTOML(`
					[[processes]]
						type = "some-type"
						command = "some-command"
						args = ["some-arg"]
						direct = true
						default = true
				`))
			})
		})

		context("when the api version is less than 0.6", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), []byte(`
api = "0.5"
[buildpack]
  id = "some-id"
  name = "some-name"
  version = "some-version"
  clear-env = false
`), 0600)).To(Succeed())
			})

			it("errors", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{
						Launch: packit.LaunchMetadata{
							Processes: []packit.Process{
								{
									Type:    "some-type",
									Command: "some-command",
									Args:    []string{"some-arg"},
									Direct:  true,
									Default: true,
								},
							},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("processes can only be marked as default with Buildpack API v0.6 or higher")))
			})
		})
	})

	context("when there are slices in the result", func() {
		it("persists a launch.toml", func() {
			packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
				return packit.BuildResult{
					Launch: packit.LaunchMetadata{
						Slices: []packit.Slice{
							{
								Paths: []string{"some-slice"},
							},
						},
					},
				}, nil
			}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

			contents, err := os.ReadFile(filepath.Join(layersDir, "launch.toml"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(contents)).To(MatchTOML(`
				[[slices]]
					paths = ["some-slice"]
			`))
		})
	})

	context("when there are labels in the result", func() {
		it("persists a launch.toml", func() {
			packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
				return packit.BuildResult{
					Launch: packit.LaunchMetadata{
						Labels: map[string]string{
							"some key":       "some value",
							"some-other-key": "some-other-value",
						},
					},
				}, nil
			}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

			contents, err := os.ReadFile(filepath.Join(layersDir, "launch.toml"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(contents)).To(MatchTOML(`
				[[labels]]
					key = "some key"
					value = "some value"

				[[labels]]
					key = "some-other-key"
					value = "some-other-value"
			`))
		})
	})

	context("when there are no processes, slices, bom or labels in the result", func() {
		it("does not persist a launch.toml", func() {
			packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
				return packit.BuildResult{}, nil
			}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

			Expect(filepath.Join(layersDir, "launch.toml")).NotTo(BeARegularFile())
		})
	})

	context("persists env vars", func() {
		context("writes to shared env folder", func() {
			it("writes env vars into env directory", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{
						Layers: []packit.Layer{
							{
								Path: filepath.Join(ctx.Layers.Path, "some-layer"),
								SharedEnv: packit.Environment{
									"SOME_VAR.append":   "append-value",
									"SOME_VAR.default":  "default-value",
									"SOME_VAR.delim":    "delim-value",
									"SOME_VAR.prepend":  "prepend-value",
									"SOME_VAR.override": "override-value",
								},
							},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

				for _, modifier := range []string{"append", "default", "delim", "prepend", "override"} {
					contents, err := os.ReadFile(filepath.Join(layersDir, "some-layer", "env", fmt.Sprintf("SOME_VAR.%s", modifier)))
					Expect(err).NotTo(HaveOccurred())
					Expect(string(contents)).To(Equal(fmt.Sprintf("%s-value", modifier)))
				}
			})
		})

		context("writes to launch folder", func() {
			it("writes env vars into env.launch directory", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{
						Layers: []packit.Layer{
							{
								Path: filepath.Join(ctx.Layers.Path, "some-layer"),
								LaunchEnv: packit.Environment{
									"SOME_VAR.append":   "append-value",
									"SOME_VAR.default":  "default-value",
									"SOME_VAR.delim":    "delim-value",
									"SOME_VAR.prepend":  "prepend-value",
									"SOME_VAR.override": "override-value",
								},
							},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

				for _, modifier := range []string{"append", "default", "delim", "prepend", "override"} {
					contents, err := os.ReadFile(filepath.Join(layersDir, "some-layer", "env.launch", fmt.Sprintf("SOME_VAR.%s", modifier)))
					Expect(err).NotTo(HaveOccurred())
					Expect(string(contents)).To(Equal(fmt.Sprintf("%s-value", modifier)))
				}
			})
			it("writes env vars into env.launch/<process> directory", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{
						Layers: []packit.Layer{
							{
								Path: filepath.Join(ctx.Layers.Path, "some-layer"),
								ProcessLaunchEnv: map[string]packit.Environment{
									"process-name": {
										"SOME_VAR.append":   "append-value",
										"SOME_VAR.default":  "default-value",
										"SOME_VAR.delim":    "delim-value",
										"SOME_VAR.prepend":  "prepend-value",
										"SOME_VAR.override": "override-value",
									},
									"another-process-name": {
										"SOME_VAR.append":   "append-value",
										"SOME_VAR.default":  "default-value",
										"SOME_VAR.delim":    "delim-value",
										"SOME_VAR.prepend":  "prepend-value",
										"SOME_VAR.override": "override-value",
									},
								},
							},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

				for _, process := range []string{"process-name", "another-process-name"} {
					for _, modifier := range []string{"append", "default", "delim", "prepend", "override"} {
						contents, err := os.ReadFile(filepath.Join(layersDir, "some-layer", "env.launch", process, fmt.Sprintf("SOME_VAR.%s", modifier)))
						Expect(err).NotTo(HaveOccurred())
						Expect(string(contents)).To(Equal(fmt.Sprintf("%s-value", modifier)))
					}
				}
			})
		})

		context("writes to build folder", func() {
			it("writes env vars into env.build directory", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{
						Layers: []packit.Layer{
							{
								Path: filepath.Join(ctx.Layers.Path, "some-layer"),
								BuildEnv: packit.Environment{
									"SOME_VAR.append":   "append-value",
									"SOME_VAR.default":  "default-value",
									"SOME_VAR.delim":    "delim-value",
									"SOME_VAR.prepend":  "prepend-value",
									"SOME_VAR.override": "override-value",
								},
							},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}))

				for _, modifier := range []string{"append", "default", "delim", "prepend", "override"} {
					contents, err := os.ReadFile(filepath.Join(layersDir, "some-layer", "env.build", fmt.Sprintf("SOME_VAR.%s", modifier)))
					Expect(err).NotTo(HaveOccurred())
					Expect(string(contents)).To(Equal(fmt.Sprintf("%s-value", modifier)))
				}
			})
		})
	})

	context("failure cases", func() {
		context("when the buildpack plan.toml is malformed", func() {
			it.Before(func() {
				err := os.WriteFile(planPath, []byte("%%%"), 0600)
				Expect(err).NotTo(HaveOccurred())
			})

			it("calls the exit handler", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("expected '.' or '=', but got '%' instead")))
			})
		})

		context("when the build func returns an error", func() {
			it("calls the exit handler", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{}, errors.New("build failed")
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError("build failed"))
			})
		})

		context("when the buildpack.toml is malformed", func() {
			it.Before(func() {
				err := os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), []byte("%%%"), 0600)
				Expect(err).NotTo(HaveOccurred())
			})

			it("calls the exit handler", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("expected '.' or '=', but got '%' instead")))
			})
		})

		context("when the buildpack plan.toml cannot be written", func() {
			it.Before(func() {
				bpTOML := []byte(`
api = "0.4"
[buildpack]
  id = "some-id"
  name = "some-name"
  version = "some-version"
  clear-env = false
				`)
				Expect(os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(envCnbDir, "buildpack.toml"), bpTOML, 0600)).To(Succeed())
				Expect(os.Chmod(planPath, 0444)).To(Succeed())
			})

			it("calls the exit handler", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{Plan: ctx.Plan}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the layer.toml file cannot be written", func() {
			it.Before(func() {
				Expect(os.Chmod(layersDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layersDir, os.ModePerm)).To(Succeed())
			})

			it("calls the exit handler", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{
						Layers: []packit.Layer{
							packit.Layer{
								Path: filepath.Join(layersDir, "some-layer"),
								Name: "some-layer",
							},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the launch.toml file cannot be written", func() {
			it.Before(func() {
				_, err := os.OpenFile(filepath.Join(layersDir, "launch.toml"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0000)
				Expect(err).NotTo(HaveOccurred())
			})

			it.After(func() {
				Expect(os.Chmod(filepath.Join(layersDir, "launch.toml"), os.ModePerm)).To(Succeed())
			})

			it("calls the exit handler", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{
						Launch: packit.LaunchMetadata{
							Processes: []packit.Process{{}},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the env dir cannot be created", func() {
			var envDir string
			it.Before(func() {
				var err error
				envDir, err = os.MkdirTemp("", "environment")
				Expect(err).NotTo(HaveOccurred())

				Expect(os.Chmod(envDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(envDir, os.ModePerm)).To(Succeed())
			})

			context("SharedEnv", func() {
				it("calls the exit handler", func() {
					packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
						return packit.BuildResult{
							Layers: []packit.Layer{
								{
									Path: envDir,
									SharedEnv: packit.Environment{
										"SOME_VAR.override": "some-value",
									},
								},
							},
						}, nil
					}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))
					Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("BuildEnv", func() {
				it("calls the exit handler", func() {
					packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
						return packit.BuildResult{
							Layers: []packit.Layer{
								{
									Path: envDir,
									BuildEnv: packit.Environment{
										"SOME_VAR.override": "some-value",
									},
								},
							},
						}, nil
					}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))
					Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("LaunchEnv", func() {
				it("calls the exit handler", func() {
					packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
						return packit.BuildResult{
							Layers: []packit.Layer{
								{
									Path: envDir,
									LaunchEnv: packit.Environment{
										"SOME_VAR.override": "some-value",
									},
								},
							},
						}, nil
					}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))
					Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})

		context("when the env file cannot be created", func() {
			context("SharedEnv", func() {
				var envDir string
				it.Before(func() {
					envDir = filepath.Join(layersDir, "some-layer", "env")
					Expect(os.MkdirAll(envDir, os.ModePerm)).To(Succeed())
					Expect(os.Chmod(envDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(envDir, os.ModePerm)).To(Succeed())
				})

				it("calls the exit handler", func() {
					packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
						return packit.BuildResult{
							Layers: []packit.Layer{{
								Path: filepath.Join(layersDir, "some-layer"),
								SharedEnv: packit.Environment{
									"SOME_VAR.override": "some-value",
								},
							}},
						}, nil
					}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))
					Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("BuildEnv", func() {
				var envDir string
				it.Before(func() {
					envDir = filepath.Join(layersDir, "some-layer", "env.build")
					Expect(os.MkdirAll(envDir, os.ModePerm)).To(Succeed())
					Expect(os.Chmod(envDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(envDir, os.ModePerm)).To(Succeed())
				})

				it("calls the exit handler", func() {
					packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
						return packit.BuildResult{
							Layers: []packit.Layer{{
								Path: filepath.Join(layersDir, "some-layer"),
								BuildEnv: packit.Environment{
									"SOME_VAR.override": "some-value",
								},
							}},
						}, nil
					}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))
					Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("LaunchEnv", func() {
				var envDir string
				it.Before(func() {
					envDir = filepath.Join(layersDir, "some-layer", "env.launch")
					Expect(os.MkdirAll(envDir, os.ModePerm)).To(Succeed())
					Expect(os.Chmod(envDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(envDir, os.ModePerm)).To(Succeed())
				})

				it("calls the exit handler", func() {
					packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
						return packit.BuildResult{
							Layers: []packit.Layer{{
								Path: filepath.Join(layersDir, "some-layer"),
								LaunchEnv: packit.Environment{
									"SOME_VAR.override": "some-value",
								},
							}},
						}, nil
					}, packit.WithArgs([]string{binaryPath, layersDir, platformDir, planPath}), packit.WithExitHandler(exitHandler))
					Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
		context("when the attempted BOM checksum algorithm is not supported", func() {
			it("persists a build.toml", func() {
				_, err := packit.GetBOMChecksumAlgorithm("RANDOM-ALG")
				Expect(err).To(MatchError("failed to get supported BOM checksum algorithm: RANDOM-ALG is not valid"))
			})
		})
	})
}
