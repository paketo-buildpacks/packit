package reload_test

import (
	"os"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/reload"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testReload(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("ShouldEnableLiveReload", func() {
		context("When it should enable live reload", func() {
			context("when BP_LIVE_RELOAD_ENABLED is 'true'", func() {
				it.Before(func() {
					Expect(os.Setenv("BP_LIVE_RELOAD_ENABLED", "true")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("BP_LIVE_RELOAD_ENABLED")).To(Succeed())
				})

				it("should require watchexec at launch", func() {
					require, err := reload.ShouldEnableLiveReload()
					Expect(err).NotTo(HaveOccurred())
					Expect(require).To(Equal(true))
				})
			})

			context("when BP_LIVE_RELOAD_ENABLED is '1'", func() {
				it.Before(func() {
					Expect(os.Setenv("BP_LIVE_RELOAD_ENABLED", "1")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("BP_LIVE_RELOAD_ENABLED")).To(Succeed())
				})

				it("should require watchexec at launch", func() {
					require, err := reload.ShouldEnableLiveReload()
					Expect(err).NotTo(HaveOccurred())
					Expect(require).To(Equal(true))
				})
			})

			context("when BP_LIVE_RELOAD_ENABLED is 'T'", func() {
				it.Before(func() {
					Expect(os.Setenv("BP_LIVE_RELOAD_ENABLED", "T")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("BP_LIVE_RELOAD_ENABLED")).To(Succeed())
				})

				it("should require watchexec at launch", func() {
					require, err := reload.ShouldEnableLiveReload()
					Expect(err).NotTo(HaveOccurred())
					Expect(require).To(Equal(true))
				})
			})
		})

		context("when it should not enable live reload", func() {
			context("when BP_LIVE_RELOAD_ENABLED is not set", func() {
				it("should not require anything at launch", func() {
					require, err := reload.ShouldEnableLiveReload()
					Expect(require).To(Equal(false))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			context("when BP_LIVE_RELOAD_ENABLED is 'F'", func() {
				it.Before(func() {
					Expect(os.Setenv("BP_LIVE_RELOAD_ENABLED", "F")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("BP_LIVE_RELOAD_ENABLED")).To(Succeed())
				})

				it("should not require anything at launch", func() {
					require, err := reload.ShouldEnableLiveReload()
					Expect(require).To(Equal(false))
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		context("failure cases", func() {
			context("when BP_LIVE_RELOAD_ENABLED is not a valid boolean value", func() {
				it.Before(func() {
					Expect(os.Setenv("BP_LIVE_RELOAD_ENABLED", "hi")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("BP_LIVE_RELOAD_ENABLED")).To(Succeed())
				})

				it("should return an error", func() {
					_, err := reload.ShouldEnableLiveReload()
					Expect(err.Error()).To(Equal(`failed to parse BP_LIVE_RELOAD_ENABLED value hi: strconv.ParseBool: parsing "hi": invalid syntax`))
				})
			})
		})
	})

	context("TransformReloadableProcesses", func() {
		var (
			originalProcess       packit.Process
			reloadableProcessSpec reload.ReloadableProcessSpec

			expectecdNonReloadableProcess packit.Process
			expectedReloadableProcess     packit.Process
		)

		it.Before(func() {
			reloadableProcessSpec = reload.ReloadableProcessSpec{
				WatchPaths: []string{
					"watch-path0",
					"watch-path1",
				},
				IgnorePaths: []string{
					"ignore-path0",
					"ignore-path1",
				},
				Shell:          "my-shell",
				VerbosityLevel: 0,
			}

			originalProcess = packit.Process{
				Type:    "my-type",
				Command: "my-command",
				Args: []string{
					"original-arg0",
					"original-arg1",
				},
				Direct:           false,
				Default:          false,
				WorkingDirectory: "my-working-directory",
			}

			expectecdNonReloadableProcess = packit.Process{
				Type:    "my-type",
				Command: "my-command",
				Args: []string{
					"original-arg0",
					"original-arg1",
				},
				Direct:           false,
				Default:          false,
				WorkingDirectory: "my-working-directory",
			}

			expectedReloadableProcess = packit.Process{
				Type:    "reload-my-type",
				Command: "watchexec",
				Args: []string{
					"--restart",
					"--watch", "watch-path0",
					"--watch", "watch-path1",
					"--ignore", "ignore-path0",
					"--ignore", "ignore-path1",
					"--shell", "my-shell",
					"--",
					"my-command",
					"original-arg0",
					"original-arg1",
				},
				Direct:           false,
				Default:          false,
				WorkingDirectory: "my-working-directory",
			}
		})

		it("should transform the original process into non-reloadable and reloadable processes", func() {
			nonReloadableProcess, reloadableProcess := reload.TransformReloadableProcesses(originalProcess, reloadableProcessSpec)

			Expect(nonReloadableProcess).To(Equal(expectecdNonReloadableProcess))
			Expect(reloadableProcess).To(Equal(expectedReloadableProcess))
		})

		context("when WatchPaths is empty", func() {
			it.Before(func() {
				reloadableProcessSpec.WatchPaths = []string{}
			})

			it("should not contain --watch args", func() {
				expectedReloadableProcess.Args = []string{
					"--restart",
					"--ignore", "ignore-path0",
					"--ignore", "ignore-path1",
					"--shell", "my-shell",
					"--",
					"my-command",
					"original-arg0",
					"original-arg1",
				}

				nonReloadableProcess, reloadableProcess := reload.TransformReloadableProcesses(originalProcess, reloadableProcessSpec)

				Expect(nonReloadableProcess).To(Equal(expectecdNonReloadableProcess))
				Expect(reloadableProcess).To(Equal(expectedReloadableProcess))
			})
		})

		context("when IgnorePaths is empty", func() {
			it.Before(func() {
				reloadableProcessSpec.IgnorePaths = []string{}
			})

			it("should not contain --ignore args", func() {
				expectedReloadableProcess.Args = []string{
					"--restart",
					"--watch", "watch-path0",
					"--watch", "watch-path1",
					"--shell", "my-shell",
					"--",
					"my-command",
					"original-arg0",
					"original-arg1",
				}

				nonReloadableProcess, reloadableProcess := reload.TransformReloadableProcesses(originalProcess, reloadableProcessSpec)

				Expect(nonReloadableProcess).To(Equal(expectecdNonReloadableProcess))
				Expect(reloadableProcess).To(Equal(expectedReloadableProcess))
			})
		})

		context("when direct and default are false", func() {
			it.Before(func() {
				originalProcess.Direct = false
				originalProcess.Default = false
			})

			it("should return indirect and non-default processes", func() {
				nonReloadableProcess, reloadableProcess := reload.TransformReloadableProcesses(originalProcess, reloadableProcessSpec)

				Expect(nonReloadableProcess.Direct).To(Equal(false))
				Expect(nonReloadableProcess.Default).To(Equal(false))

				Expect(reloadableProcess.Direct).To(Equal(false))
				Expect(reloadableProcess.Default).To(Equal(false))
			})
		})

		context("when default is true", func() {
			it.Before(func() {
				originalProcess.Direct = false
				originalProcess.Default = true
			})

			it("should return a default reloadable process", func() {
				nonReloadableProcess, reloadableProcess := reload.TransformReloadableProcesses(originalProcess, reloadableProcessSpec)

				Expect(nonReloadableProcess.Direct).To(Equal(false))
				Expect(nonReloadableProcess.Default).To(Equal(false))

				Expect(reloadableProcess.Direct).To(Equal(false))
				Expect(reloadableProcess.Default).To(Equal(true))
			})
		})

		context("when direct is true", func() {
			it.Before(func() {
				originalProcess.Direct = true
				originalProcess.Default = false
			})

			it("should return direct processes", func() {
				nonReloadableProcess, reloadableProcess := reload.TransformReloadableProcesses(originalProcess, reloadableProcessSpec)

				Expect(nonReloadableProcess.Direct).To(Equal(true))
				Expect(nonReloadableProcess.Default).To(Equal(false))

				Expect(reloadableProcess.Direct).To(Equal(true))
				Expect(reloadableProcess.Default).To(Equal(false))
			})
		})

		context("when direct and default are true", func() {
			it.Before(func() {
				originalProcess.Direct = true
				originalProcess.Default = true
			})

			it("should return direct processes, with a default reloadable process", func() {
				nonReloadableProcess, reloadableProcess := reload.TransformReloadableProcesses(originalProcess, reloadableProcessSpec)

				Expect(nonReloadableProcess.Direct).To(Equal(true))
				Expect(nonReloadableProcess.Default).To(Equal(false))

				Expect(reloadableProcess.Direct).To(Equal(true))
				Expect(reloadableProcess.Default).To(Equal(true))
			})
		})

		context("when spec.VerbosityLevel is 1", func() {
			it.Before(func() {
				reloadableProcessSpec.VerbosityLevel = 1
			})

			it("should return args with 1 verbose flag", func() {
				expectedReloadableProcess.Args = []string{
					"--restart",
					"--watch", "watch-path0",
					"--watch", "watch-path1",
					"--ignore", "ignore-path0",
					"--ignore", "ignore-path1",
					"--shell", "my-shell",
					"-v",
					"--",
					"my-command",
					"original-arg0",
					"original-arg1",
				}

				nonReloadableProcess, reloadableProcess := reload.TransformReloadableProcesses(originalProcess, reloadableProcessSpec)

				Expect(nonReloadableProcess).To(Equal(expectecdNonReloadableProcess))
				Expect(reloadableProcess).To(Equal(expectedReloadableProcess))
			})
		})

		context("when spec.VerbosityLevel is 4", func() {
			it.Before(func() {
				reloadableProcessSpec.VerbosityLevel = 4
			})

			it("should return args with 4 verbose flag", func() {
				expectedReloadableProcess.Args = []string{
					"--restart",
					"--watch", "watch-path0",
					"--watch", "watch-path1",
					"--ignore", "ignore-path0",
					"--ignore", "ignore-path1",
					"--shell", "my-shell",
					"-vvvv",
					"--",
					"my-command",
					"original-arg0",
					"original-arg1",
				}

				nonReloadableProcess, reloadableProcess := reload.TransformReloadableProcesses(originalProcess, reloadableProcessSpec)

				Expect(nonReloadableProcess).To(Equal(expectecdNonReloadableProcess))
				Expect(reloadableProcess).To(Equal(expectedReloadableProcess))
			})
		})

		context("when spec.Shell is empty", func() {
			it.Before(func() {
				reloadableProcessSpec.Shell = ""
			})

			it("will default to --shell=none", func() {
				expectedReloadableProcess.Args = []string{
					"--restart",
					"--watch", "watch-path0",
					"--watch", "watch-path1",
					"--ignore", "ignore-path0",
					"--ignore", "ignore-path1",
					"--shell", "none",
					"--",
					"my-command",
					"original-arg0",
					"original-arg1",
				}

				nonReloadableProcess, reloadableProcess := reload.TransformReloadableProcesses(originalProcess, reloadableProcessSpec)

				Expect(nonReloadableProcess).To(Equal(expectecdNonReloadableProcess))
				Expect(reloadableProcess).To(Equal(expectedReloadableProcess))
			})
		})
	})

	context("WatchExecRequirement", func() {
		it("should require watchexec at launch", func() {
			Expect(reload.WatchExecRequirement).To(Equal(packit.BuildPlanRequirement{
				Name: "watchexec",
				Metadata: map[string]interface{}{
					"launch": true,
				}}))
		})
	})
}
