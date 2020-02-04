package packit_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/fakes"
	"github.com/cloudfoundry/packit/internal"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/packit/matchers"
	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir  string
		tmpDir      string
		cnbDir      string
		binaryPath  string
		exitHandler *fakes.ExitHandler
	)

	it.Before(func() {
		var err error
		workingDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		tmpDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		tmpDir, err = filepath.EvalSymlinks(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		Expect(os.Chdir(tmpDir)).To(Succeed())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		binaryPath = filepath.Join(cnbDir, "bin", "detect")

		Expect(ioutil.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), []byte(`
[buildpack]
id = "some-id"
name = "some-name"
version = "some-version"
clear-env = false
`), 0644)).To(Succeed())

		exitHandler = &fakes.ExitHandler{}
	})

	it.After(func() {
		Expect(os.Chdir(workingDir)).To(Succeed())
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
	})

	context("when providing the detect context to the given DetectFunc", func() {
		var filePath string

		it.Before(func() {
			filePath = filepath.Join(os.TempDir(), "buildplan.toml")
		})

		it("succeeds", func() {
			var context packit.DetectContext

			packit.Detect(func(ctx packit.DetectContext) (packit.DetectResult, error) {
				context = ctx

				return packit.DetectResult{}, nil
			}, packit.WithArgs([]string{binaryPath, "", filePath}))

			Expect(context).To(Equal(packit.DetectContext{
				WorkingDir: tmpDir,
				BuildpackInfo: packit.BuildpackInfo{
					ID:      "some-id",
					Name:    "some-name",
					Version: "some-version",
				},
			}))
		})
	})

	it("writes out the buildplan.toml", func() {
		path := filepath.Join(tmpDir, "buildplan.toml")

		packit.Detect(func(packit.DetectContext) (packit.DetectResult, error) {
			return packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: "some-provision"},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name:    "some-requirement",
							Version: "some-version",
							Metadata: map[string]string{
								"some-key": "some-value",
							},
						},
					},
				},
			}, nil
		}, packit.WithArgs([]string{binaryPath, "", path}))

		contents, err := ioutil.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(contents)).To(MatchTOML(`
[[provides]]
name = "some-provision"

[[requires]]
name = "some-requirement"
version = "some-version"

[requires.metadata]
some-key = "some-value"
`))
	})

	context("when the DetectFunc returns an error", func() {
		it("calls the ExitHandler with that error", func() {
			packit.Detect(func(ctx packit.DetectContext) (packit.DetectResult, error) {
				return packit.DetectResult{}, errors.New("failed to detect")
			}, packit.WithArgs([]string{binaryPath, "", ""}), packit.WithExitHandler(exitHandler))

			Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError("failed to detect"))
		})
	})

	context("when the DetectFunc fails", func() {
		it("calls the ExitHandler with the correct exit code", func() {
			var exitCode int

			packit.Detect(func(ctx packit.DetectContext) (packit.DetectResult, error) {
				return packit.DetectResult{}, packit.Fail
			}, packit.WithArgs([]string{binaryPath, "", ""}), packit.WithExitHandler(internal.NewExitHandler(internal.WithExitHandlerExitFunc(func(code int) {
				exitCode = code
			}))))

			Expect(exitCode).To(Equal(100))
		})
	})

	context("failure cases", func() {
		context("when the buildpack.toml cannot be read", func() {
			it("returns an error", func() {
				path := filepath.Join(tmpDir, "buildplan.toml")

				packit.Detect(func(packit.DetectContext) (packit.DetectResult, error) {
					return packit.DetectResult{}, nil
				}, packit.WithArgs([]string{"", "", path}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("no such file or directory")))
			})
		})

		context("when the buildplan.toml cannot be opened", func() {
			it("returns an error", func() {
				path := filepath.Join(tmpDir, "buildplan.toml")
				_, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0000)
				Expect(err).NotTo(HaveOccurred())

				packit.Detect(func(packit.DetectContext) (packit.DetectResult, error) {
					return packit.DetectResult{
						Plan: packit.BuildPlan{
							Provides: []packit.BuildPlanProvision{
								{Name: "some-provision"},
							},
							Requires: []packit.BuildPlanRequirement{
								{
									Name:    "some-requirement",
									Version: "some-version",
									Metadata: map[string]string{
										"some-key": "some-value",
									},
								},
							},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, "", path}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the buildplan.toml cannot be encoded", func() {
			it("returns an error", func() {
				path := filepath.Join(tmpDir, "buildplan.toml")

				packit.Detect(func(packit.DetectContext) (packit.DetectResult, error) {
					return packit.DetectResult{
						Plan: packit.BuildPlan{
							Provides: []packit.BuildPlanProvision{
								{Name: "some-provision"},
							},
							Requires: []packit.BuildPlanRequirement{
								{
									Name:     "some-requirement",
									Version:  "some-version",
									Metadata: map[int]int{},
								},
							},
						},
					}, nil
				}, packit.WithArgs([]string{binaryPath, "", path}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("cannot encode a map with non-string key type")))
			})
		})
	})
}
