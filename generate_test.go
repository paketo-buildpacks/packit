package packit_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testGenerate(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir  string
		platformDir string
		tmpDir      string
		planPath    string
		cnbDir      string
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

		extTOML := []byte(`
api = "0.9"
[extension]
  id = "some-id"
  name = "some-name"
  version = "some-version"
  homepage = "some-homepage"
  description = "some-description"
  keywords = ["some-keyword"]

  [[extension.licenses]]
	type = "some-license-type"
	uri = "some-license-uri"
`)
		Expect(os.WriteFile(filepath.Join(cnbDir, "extension.toml"), extTOML, 0600)).To(Succeed())

		binaryPath = filepath.Join(cnbDir, "bin", "generate")

		Expect(os.Setenv("CNB_STACK_ID", "some-stack")).To(Succeed())
		Expect(os.Setenv("CNB_BP_PLAN_PATH", planPath)).To(Succeed())
		Expect(os.Setenv("CNB_PLATFORM_DIR", platformDir)).To(Succeed())
		Expect(os.Setenv("CNB_EXTENSION_DIR", cnbDir)).To(Succeed())

		exitHandler = &fakes.ExitHandler{}
		exitHandler.ErrorCall.Stub = func(err error) {
			Expect(err).NotTo(HaveOccurred())
		}

	})

	it.After(func() {
		Expect(os.Unsetenv("CNB_STACK_ID")).To(Succeed())
		Expect(os.Unsetenv("CNB_BP_PLAN_PATH")).To(Succeed())
		Expect(os.Unsetenv("CNB_PLATFORM_DIR")).To(Succeed())
		Expect(os.Unsetenv("CNB_EXTENSION_DIR")).To(Succeed())

		Expect(os.Chdir(workingDir)).To(Succeed())
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
		Expect(os.RemoveAll(platformDir)).To(Succeed())
	})

	it("provides the generate context to the given GenerateFunc", func() {
		var context packit.GenerateContext
		packit.Generate(func(ctx packit.GenerateContext) (packit.GenerateResult, error) {
			context = ctx

			return packit.GenerateResult{}, nil
		}, packit.WithArgs([]string{binaryPath}), packit.WithExitHandler(exitHandler))

		Expect(exitHandler.ErrorCall.CallCount).To(Equal(0))
		Expect(context).To(Equal(packit.GenerateContext{
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
			Info: packit.Info{
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

	context("failure cases", func() {
		context("when the buildpack plan.toml is malformed", func() {
			it.Before(func() {
				err := os.WriteFile(planPath, []byte("%%%"), 0600)
				Expect(err).NotTo(HaveOccurred())
				exitHandler.ErrorCall.Stub = nil
			})

			it("calls the exit handler", func() {
				packit.Generate(func(ctx packit.GenerateContext) (packit.GenerateResult, error) {
					return packit.GenerateResult{}, nil
				}, packit.WithArgs([]string{binaryPath}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("expected '.' or '=', but got '%' instead")))
			})
		})

		context("when the generate func returns an error", func() {
			it("calls the exit handler", func() {
				exitHandler.ErrorCall.Stub = nil
				packit.Generate(func(ctx packit.GenerateContext) (packit.GenerateResult, error) {
					return packit.GenerateResult{}, errors.New("generate failed")
				}, packit.WithArgs([]string{binaryPath}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError("generate failed"))
			})
		})

		context("when the exension.toml is malformed", func() {
			it.Before(func() {
				err := os.WriteFile(filepath.Join(cnbDir, "extension.toml"), []byte("%%%"), 0600)
				Expect(err).NotTo(HaveOccurred())
				exitHandler.ErrorCall.Stub = nil
			})

			it("calls the exit handler", func() {
				packit.Generate(func(ctx packit.GenerateContext) (packit.GenerateResult, error) {
					return packit.GenerateResult{}, nil
				}, packit.WithArgs([]string{binaryPath}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("expected '.' or '=', but got '%' instead")))
			})
		})
	})
}
