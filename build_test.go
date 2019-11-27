package packit_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir  string
		tmpDir      string
		layersDir   string
		planPath    string
		exitHandler *fakes.ExitHandler
	)

	it.Before(func() {
		var err error
		workingDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		tmpDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		tmpDir, err = filepath.EvalSymlinks(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		Expect(os.Chdir(tmpDir)).To(Succeed())

		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		file, err := ioutil.TempFile("", "plan.toml")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		_, err = file.WriteString(`
[[entries]]
name = "some-entry"
version = "some-version"

[entries.metadata]
some-key = "some-value"
`)
		Expect(err).NotTo(HaveOccurred())

		planPath = file.Name()

		Expect(os.Setenv("CNB_STACK_ID", "some-stack")).To(Succeed())

		exitHandler = &fakes.ExitHandler{}
	})

	it.After(func() {
		Expect(os.Unsetenv("CNB_STACK_ID")).To(Succeed())

		Expect(os.Chdir(workingDir)).To(Succeed())
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
		Expect(os.RemoveAll(layersDir)).To(Succeed())
	})

	it("provides the build context to the given BuildFunc", func() {
		var context packit.BuildContext

		packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
			context = ctx

			return packit.BuildResult{}, nil
		}, packit.WithArgs([]string{"/cnbs/some-cnb/bin/build", layersDir, "", planPath}))

		Expect(context).To(Equal(packit.BuildContext{
			CNBPath:    "/cnbs/some-cnb",
			Stack:      "some-stack",
			WorkingDir: tmpDir,
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name:    "some-entry",
						Version: "some-version",
						Metadata: map[string]interface{}{
							"some-key": "some-value",
						},
					},
				},
			},
			Layers: packit.Layers{
				Path: layersDir,
			},
		}))
	})

	it("updates the buildpack plan.toml with any changes", func() {
		packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
			ctx.Plan.Entries[0].Metadata["other-key"] = "other-value"

			return packit.BuildResult{
				Plan: ctx.Plan,
			}, nil
		}, packit.WithArgs([]string{"", "", "", planPath}))

		contents, err := ioutil.ReadFile(planPath)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(contents)).To(MatchTOML(`
[[entries]]
name = "some-entry"
version = "some-version"

[entries.metadata]
some-key = "some-value"
other-key = "other-value"
`))
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
		}, packit.WithArgs([]string{"", layersDir, "", planPath}))

		contents, err := ioutil.ReadFile(filepath.Join(layersDir, "some-layer.toml"))
		Expect(err).NotTo(HaveOccurred())

		Expect(string(contents)).To(MatchTOML(`
launch = true
build = true
cache = true

[metadata]
some-key = "some-value"
`))
	})

	it("persists a launch.toml", func() {
		packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
			return packit.BuildResult{
				Processes: []packit.Process{
					{
						Type:    "some-type",
						Command: "some-command",
						Args:    []string{"some-arg"},
						Direct:  true,
					},
				},
			}, nil
		}, packit.WithArgs([]string{"", layersDir, "", planPath}))

		contents, err := ioutil.ReadFile(filepath.Join(layersDir, "launch.toml"))
		Expect(err).NotTo(HaveOccurred())

		Expect(string(contents)).To(MatchTOML(`
[[processes]]
type = "some-type"
command = "some-command"
args = ["some-arg"]
direct = true
`))
	})

	context("when there are no processes in the result", func() {
		it("does not persist a launch.toml", func() {
			packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
				return packit.BuildResult{}, nil
			}, packit.WithArgs([]string{"", layersDir, "", planPath}))

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
				}, packit.WithArgs([]string{"", layersDir, "", planPath}))

				for _, modifier := range []string{"append", "default", "delim", "prepend", "override"} {
					contents, err := ioutil.ReadFile(filepath.Join(layersDir, "some-layer", "env", fmt.Sprintf("SOME_VAR.%s", modifier)))
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
				}, packit.WithArgs([]string{"", layersDir, "", planPath}))

				for _, modifier := range []string{"append", "default", "delim", "prepend", "override"} {
					contents, err := ioutil.ReadFile(filepath.Join(layersDir, "some-layer", "env.launch", fmt.Sprintf("SOME_VAR.%s", modifier)))
					Expect(err).NotTo(HaveOccurred())
					Expect(string(contents)).To(Equal(fmt.Sprintf("%s-value", modifier)))
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
				}, packit.WithArgs([]string{"", layersDir, "", planPath}))

				for _, modifier := range []string{"append", "default", "delim", "prepend", "override"} {
					contents, err := ioutil.ReadFile(filepath.Join(layersDir, "some-layer", "env.build", fmt.Sprintf("SOME_VAR.%s", modifier)))
					Expect(err).NotTo(HaveOccurred())
					Expect(string(contents)).To(Equal(fmt.Sprintf("%s-value", modifier)))
				}
			})
		})
	})

	context("failure cases", func() {
		context("when the buildpack plan.toml is malformed", func() {
			it.Before(func() {
				err := ioutil.WriteFile(planPath, []byte("%%%"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("calls the exit handler", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{}, nil
				}, packit.WithArgs([]string{"", "", "", planPath}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("bare keys cannot contain '%'")))
			})
		})

		context("when the build func returns an error", func() {
			it("calls the exit handler", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{}, errors.New("build failed")
				}, packit.WithArgs([]string{"", "", "", planPath}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError("build failed"))
			})
		})

		context("when the buildpack plan.toml cannot be written", func() {
			it.Before(func() {
				Expect(os.Chmod(planPath, 0444)).To(Succeed())
			})

			it("calls the exit handler", func() {
				packit.Build(func(ctx packit.BuildContext) (packit.BuildResult, error) {
					return packit.BuildResult{}, nil
				}, packit.WithArgs([]string{"", "", "", planPath}), packit.WithExitHandler(exitHandler))

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
				}, packit.WithArgs([]string{"", layersDir, "", planPath}), packit.WithExitHandler(exitHandler))

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
						Processes: []packit.Process{{}},
					}, nil
				}, packit.WithArgs([]string{"", layersDir, "", planPath}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the env dir cannot be created", func() {
			var envDir string
			it.Before(func() {
				var err error
				envDir, err = ioutil.TempDir("", "environment")
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
					}, packit.WithArgs([]string{"", layersDir, "", planPath}), packit.WithExitHandler(exitHandler))
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
					}, packit.WithArgs([]string{"", layersDir, "", planPath}), packit.WithExitHandler(exitHandler))
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
					}, packit.WithArgs([]string{"", layersDir, "", planPath}), packit.WithExitHandler(exitHandler))
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
					}, packit.WithArgs([]string{"", layersDir, "", planPath}), packit.WithExitHandler(exitHandler))
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
					}, packit.WithArgs([]string{"", layersDir, "", planPath}), packit.WithExitHandler(exitHandler))
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
					}, packit.WithArgs([]string{"", layersDir, "", planPath}), packit.WithExitHandler(exitHandler))
					Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
}
