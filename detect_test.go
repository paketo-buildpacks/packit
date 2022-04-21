package packit_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fakes"
	"github.com/paketo-buildpacks/packit/v2/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/packit/v2/matchers"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir  string
		tmpDir      string
		platformDir string
		cnbDir      string
		binaryPath  string
		stackID     string
		planDir     string
		planPath    string

		exitHandler *fakes.ExitHandler
	)

	it.Before(func() {
		var err error
		workingDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		tmpDir, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())

		tmpDir, err = filepath.EvalSymlinks(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		Expect(os.Chdir(tmpDir)).To(Succeed())

		platformDir, err = os.MkdirTemp("", "platform")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		stackID = "io.packit.test.stack"
		Expect(os.Setenv("CNB_STACK_ID", stackID)).To(Succeed())

		binaryPath = filepath.Join(cnbDir, "bin", "detect")

		bpTOMLContent := []byte(`
api = "0.5"
[buildpack]
  id = "some-id"
  name = "some-name"
  version = "some-version"
  clear-env = false
`)
		Expect(os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), bpTOMLContent, 0600)).To(Succeed())

		planDir, err = os.MkdirTemp("", "buildplan.toml")
		Expect(err).NotTo(HaveOccurred())

		planPath = filepath.Join(planDir, "buildplan.toml")

		exitHandler = &fakes.ExitHandler{}
	})

	it.After(func() {
		Expect(os.Chdir(workingDir)).To(Succeed())
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(planDir)).To(Succeed())
		Expect(os.RemoveAll(platformDir)).To(Succeed())
		Expect(os.Unsetenv("CNB_STACK_ID")).To(Succeed())
	})

	context("when providing the detect context to the given DetectFunc", func() {
		it("succeeds", func() {
			var context packit.DetectContext

			packit.Detect(func(ctx packit.DetectContext) (packit.DetectResult, error) {
				context = ctx

				return packit.DetectResult{}, nil
			}, packit.WithArgs([]string{binaryPath, platformDir, planPath}))

			Expect(context).To(Equal(packit.DetectContext{
				WorkingDir: tmpDir,
				CNBPath:    cnbDir,
				Platform: packit.Platform{
					Path: platformDir,
				},
				BuildpackInfo: packit.BuildpackInfo{
					ID:      "some-id",
					Name:    "some-name",
					Version: "some-version",
				},
				Stack: stackID,
			}))
		})
	})

	it("writes out the buildplan.toml", func() {
		packit.Detect(func(packit.DetectContext) (packit.DetectResult, error) {
			return packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: "some-provision"},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: "some-requirement",
							Metadata: map[string]string{
								"version":  "some-version",
								"some-key": "some-value",
							},
						},
					},
				},
			}, nil
		}, packit.WithArgs([]string{binaryPath, platformDir, planPath}))

		contents, err := os.ReadFile(planPath)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(contents)).To(MatchTOML(`
[[provides]]
  name = "some-provision"

[[requires]]
  name = "some-requirement"

[requires.metadata]
  version = "some-version"
  some-key = "some-value"
`))
	})

	it("writes out the buildplan.toml with multiple plans", func() {
		packit.Detect(func(packit.DetectContext) (packit.DetectResult, error) {
			return packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: "some-provision"},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: "some-requirement",
							Metadata: map[string]string{
								"version":  "some-version",
								"some-key": "some-value",
							},
						},
					},
					Or: []packit.BuildPlan{
						{
							Provides: []packit.BuildPlanProvision{
								{Name: "some-other-provision"},
							},
							Requires: []packit.BuildPlanRequirement{
								{
									Name: "some-other-requirement",
									Metadata: map[string]string{
										"version":        "some-other-version",
										"some-other-key": "some-other-value",
									},
								},
							},
						},
						{
							Provides: []packit.BuildPlanProvision{
								{Name: "some-another-provision"},
							},
							Requires: []packit.BuildPlanRequirement{
								{
									Name: "some-another-requirement",
									Metadata: map[string]string{
										"version":          "some-another-version",
										"some-another-key": "some-another-value",
									},
								},
							},
						},
					},
				},
			}, nil
		}, packit.WithArgs([]string{binaryPath, platformDir, planPath}))

		contents, err := os.ReadFile(planPath)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(contents)).To(MatchTOML(`
[[provides]]
  name = "some-provision"

[[requires]]
  name = "some-requirement"

  [requires.metadata]
  version = "some-version"
	some-key = "some-value"

[[or]]

  [[or.provides]]
	name = "some-other-provision"

  [[or.requires]]
	name = "some-other-requirement"

	[or.requires.metadata]
		version = "some-other-version"
		some-other-key = "some-other-value"

[[or]]

  [[or.provides]]
	name = "some-another-provision"

  [[or.requires]]
	name = "some-another-requirement"
	[or.requires.metadata]
		version = "some-another-version"
	  some-another-key = "some-another-value"
`))
	})

	context("when CNB_BUILDPACK_DIR is set", func() {
		it.Before(func() {
			Expect(os.Setenv("CNB_BUILDPACK_DIR", cnbDir)).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("CNB_BUILDPACK_DIR")).To(Succeed())
		})

		it("the Detect context receives the correct value", func() {
			var context packit.DetectContext

			packit.Detect(func(ctx packit.DetectContext) (packit.DetectResult, error) {
				context = ctx

				return packit.DetectResult{}, nil
			}, packit.WithArgs([]string{"env-var-override", platformDir, planPath}))

			Expect(context).To(Equal(packit.DetectContext{
				WorkingDir: tmpDir,
				CNBPath:    cnbDir,
				Platform: packit.Platform{
					Path: platformDir,
				},
				BuildpackInfo: packit.BuildpackInfo{
					ID:      "some-id",
					Name:    "some-name",
					Version: "some-version",
				},
				Stack: stackID,
			}))
		})
	})

	context("when CNB_PLATFORM_DIR is set", func() {
		it.Before(func() {
			Expect(os.Setenv("CNB_PLATFORM_DIR", platformDir)).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("CNB_PLATFORM_DIR")).To(Succeed())
		})

		it("the Detect context receives the correct value", func() {
			var context packit.DetectContext

			packit.Detect(func(ctx packit.DetectContext) (packit.DetectResult, error) {
				context = ctx

				return packit.DetectResult{}, nil
			}, packit.WithArgs([]string{binaryPath, "env-var-override", planPath}))

			Expect(context).To(Equal(packit.DetectContext{
				WorkingDir: tmpDir,
				CNBPath:    cnbDir,
				Platform: packit.Platform{
					Path: platformDir,
				},
				BuildpackInfo: packit.BuildpackInfo{
					ID:      "some-id",
					Name:    "some-name",
					Version: "some-version",
				},
				Stack: stackID,
			}))
		})
	})

	context("when CNB_BUILD_PLAN_PATH is set", func() {
		it.Before(func() {
			Expect(os.Setenv("CNB_BUILD_PLAN_PATH", planPath)).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("CNB_BUILD_PLAN_PATH")).To(Succeed())
		})

		it("the Detect context receives the correct value", func() {
			var context packit.DetectContext

			packit.Detect(func(ctx packit.DetectContext) (packit.DetectResult, error) {
				context = ctx

				return packit.DetectResult{
					Plan: packit.BuildPlan{
						Provides: []packit.BuildPlanProvision{
							{Name: "some-provision"},
						},
						Requires: []packit.BuildPlanRequirement{
							{
								Name: "some-requirement",
								Metadata: map[string]string{
									"version":  "some-version",
									"some-key": "some-value",
								},
							},
						},
					},
				}, nil
			}, packit.WithArgs([]string{binaryPath, platformDir, "env-var-override"}))

			Expect(context).To(Equal(packit.DetectContext{
				WorkingDir: tmpDir,
				CNBPath:    cnbDir,
				Platform: packit.Platform{
					Path: platformDir,
				},
				BuildpackInfo: packit.BuildpackInfo{
					ID:      "some-id",
					Name:    "some-name",
					Version: "some-version",
				},
				Stack: stackID,
			}))

			contents, err := os.ReadFile(planPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(contents)).To(MatchTOML(`
[[provides]]
  name = "some-provision"

[[requires]]
  name = "some-requirement"

[requires.metadata]
  version = "some-version"
  some-key = "some-value"
`))
		})
	})

	context("when the DetectFunc returns an error", func() {
		it("calls the ExitHandler with that error", func() {
			packit.Detect(func(ctx packit.DetectContext) (packit.DetectResult, error) {
				return packit.DetectResult{}, errors.New("failed to detect")
			}, packit.WithArgs([]string{binaryPath, platformDir, planPath}), packit.WithExitHandler(exitHandler))

			Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError("failed to detect"))
		})
	})

	context("when the DetectFunc fails", func() {
		it("calls the ExitHandler with the correct exit code", func() {
			var exitCode int
			buffer := bytes.NewBuffer(nil)

			packit.Detect(func(ctx packit.DetectContext) (packit.DetectResult, error) {
				return packit.DetectResult{}, packit.Fail.WithMessage("failure message")
			},
				packit.WithArgs([]string{binaryPath, platformDir, planPath}),
				packit.WithExitHandler(
					internal.NewExitHandler(
						internal.WithExitHandlerExitFunc(func(code int) {
							exitCode = code
						}),
						internal.WithExitHandlerStderr(buffer),
					),
				),
			)

			Expect(exitCode).To(Equal(100))
			Expect(buffer.String()).To(Equal("failure message\n"))
		})
	})

	context("failure cases", func() {
		context("when the buildpack.toml cannot be read", func() {
			it("returns an error", func() {
				packit.Detect(func(packit.DetectContext) (packit.DetectResult, error) {
					return packit.DetectResult{}, nil
				}, packit.WithArgs([]string{binaryPath, platformDir, "/no/such/plan/path"}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("no such file or directory")))
			})
		})

		context("when the buildplan.toml cannot be opened", func() {
			it.Before(func() {
				_, err := os.OpenFile(planPath, os.O_CREATE|os.O_RDWR, 0000)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns an error", func() {
				packit.Detect(func(packit.DetectContext) (packit.DetectResult, error) {
					return packit.DetectResult{
						Plan: packit.BuildPlan{
							Provides: []packit.BuildPlanProvision{
								{Name: "some-provision"},
							},
							Requires: []packit.BuildPlanRequirement{
								{
									Name: "some-requirement",
									Metadata: map[string]string{
										"version":  "some-version",
										"some-key": "some-value",
									},
								},
							},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, platformDir, planPath}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the buildplan.toml cannot be encoded", func() {
			it("returns an error", func() {
				packit.Detect(func(packit.DetectContext) (packit.DetectResult, error) {
					return packit.DetectResult{
						Plan: packit.BuildPlan{
							Provides: []packit.BuildPlanProvision{
								{Name: "some-provision"},
							},
							Requires: []packit.BuildPlanRequirement{
								{
									Name:     "some-requirement",
									Metadata: map[int]int{},
								},
							},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, platformDir, planPath}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("cannot encode a map with non-string key type")))
			})
		})
	})
}
