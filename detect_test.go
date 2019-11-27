package packit_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir  string
		tmpDir      string
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

		exitHandler = &fakes.ExitHandler{}
	})

	it.After(func() {
		Expect(os.Chdir(workingDir)).To(Succeed())
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	context("when providing the detect context to the given DetectFunc", func() {
		var oldArgs []string

		// Test depends on os.Args having at least 3 elements
		it.Before(func() {
			filePath := filepath.Join(os.TempDir(), "buildpack.toml")
			oldArgs = os.Args
			os.Args = append(os.Args[:2], append([]string{filePath, filePath}, os.Args[2:]...)...)
		})

		it.After(func() {
			os.Args = oldArgs
		})

		it("succeeds", func() {
			var context packit.DetectContext

			packit.Detect(func(ctx packit.DetectContext) (packit.DetectResult, error) {
				context = ctx

				return packit.DetectResult{}, nil
			})

			Expect(context).To(Equal(packit.DetectContext{
				WorkingDir: tmpDir,
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
		}, packit.WithArgs([]string{"", "", path}))

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
			}, packit.WithExitHandler(exitHandler))

			Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError("failed to detect"))
		})
	})

	context("failure cases", func() {
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
				}, packit.WithArgs([]string{"", "", path}), packit.WithExitHandler(exitHandler))

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
				}, packit.WithArgs([]string{"", "", path}), packit.WithExitHandler(exitHandler))

				Expect(exitHandler.ErrorCall.Receives.Error).To(MatchError(ContainSubstring("cannot encode a map with non-string key type")))
			})
		})
	})
}
