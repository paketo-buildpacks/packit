package packit_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testRun(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir  string
		tmpDir      string
		cnbDir      string
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

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.WriteFile(filepath.Join(cnbDir, "buildpack.toml"), []byte(`
api = "0.5"
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

	context("when running the detect executable", func() {
		var (
			args          []string
			buildPlanPath string
		)

		it.Before(func() {
			buildPlanPath = filepath.Join(tmpDir, "buildplan.toml")

			args = []string{filepath.Join(cnbDir, "bin", "detect"), "", buildPlanPath}
		})

		it.After(func() {
			Expect(os.Remove(buildPlanPath)).To(Succeed())
		})

		it("calls the DetectFunc", func() {
			var detectCalled bool

			detect := func(packit.DetectContext) (packit.DetectResult, error) {
				detectCalled = true
				return packit.DetectResult{}, nil
			}

			packit.Run(detect, nil, packit.WithArgs(args), packit.WithExitHandler(exitHandler))

			Expect(detectCalled).To(BeTrue())
			Expect(exitHandler.ErrorCall.CallCount).To(Equal(0))
		})
	})

	context("when running the build executable", func() {
		var (
			args      []string
			layersDir string
			planPath  string
		)

		it.Before(func() {
			file, err := os.CreateTemp("", "plan.toml")
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

			layersDir, err = os.MkdirTemp("", "layers")
			Expect(err).NotTo(HaveOccurred())

			args = []string{filepath.Join(cnbDir, "bin", "build"), layersDir, "", planPath}
		})

		it.After(func() {
			Expect(os.RemoveAll(layersDir)).To(Succeed())
			Expect(os.Remove(planPath)).To(Succeed())
		})

		it("calls the BuildFunc", func() {
			var buildCalled bool

			build := func(packit.BuildContext) (packit.BuildResult, error) {
				buildCalled = true
				return packit.BuildResult{}, nil
			}

			packit.Run(nil, build, packit.WithArgs(args), packit.WithExitHandler(exitHandler))

			Expect(buildCalled).To(BeTrue())
			Expect(exitHandler.ErrorCall.CallCount).To(Equal(0))
		})
	})

	context("when running any other executable", func() {
		var args []string

		it.Before(func() {
			args = []string{filepath.Join(cnbDir, "bin", "something-else"), "some", "args"}
		})

		it("returns an error", func() {
			packit.Run(nil, nil, packit.WithArgs(args), packit.WithExitHandler(exitHandler))

			Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError("failed to run buildpack: unknown lifecycle phase \"something-else\""))
		})
	})
}
