package pexec_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega/gexec"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPexec(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		tmpDir         string
		stdout, stderr *bytes.Buffer

		executable     pexec.Executable
		executableName string
	)

	it.Before(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "executable")
		Expect(err).NotTo(HaveOccurred())

		tmpDir, err = filepath.EvalSymlinks(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		stdout = bytes.NewBuffer(nil)
		stderr = bytes.NewBuffer(nil)

		executableName = filepath.Base(fakeCLI)
		executable = pexec.NewExecutable(executableName)
	})

	it.After(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	context("Execute", func() {
		it("executes the given arguments against the executable", func() {
			err := executable.Execute(pexec.Execution{
				Args:   []string{"something"},
				Stdout: stdout,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(ContainSubstring(fmt.Sprintf("Arguments: [%s something]", fakeCLI)))
		})

		context("when given a execution directory", func() {
			it("executes within that directory", func() {
				err := executable.Execute(pexec.Execution{
					Dir:    tmpDir,
					Stdout: stdout,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(ContainSubstring(fmt.Sprintf("PWD=%s", tmpDir)))
			})
		})

		context("when given an execution environment", func() {
			it("executes with that environment", func() {
				err := executable.Execute(pexec.Execution{
					Env:    []string{"SOME_KEY=some-value"},
					Stdout: stdout,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(ContainSubstring("SOME_KEY=some-value"))
			})
		})

		context("when given a writer for stdout and stderr", func() {
			it("pipes stdout to that writer", func() {
				err := executable.Execute(pexec.Execution{
					Stdout: stdout,
					Stderr: stderr,
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(stdout).To(ContainSubstring("Output on stdout"))
				Expect(stderr).To(ContainSubstring("Output on stderr"))
			})
		})

		context("when the executable is on the PATH given as an argument", func() {
			it.Before(func() {
				os.Setenv("PATH", "some-path")
			})

			it("executes the given arguments against the executable", func() {
				err := executable.Execute(pexec.Execution{
					Args:   []string{"something"},
					Env:    []string{fmt.Sprintf("PATH=%s", filepath.Dir(fakeCLI))},
					Stdout: stdout,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(ContainSubstring(fmt.Sprintf("Arguments: [%s something]", fakeCLI)))
				Expect(os.Getenv("PATH")).To(Equal("some-path"))
			})
		})

		context("failure cases", func() {
			context("when the executable cannot be found on the path", func() {
				it.Before(func() {
					executable = pexec.NewExecutable("unknown-executable")
				})

				it("returns an error", func() {
					err := executable.Execute(pexec.Execution{})
					Expect(err).To(MatchError("exec: \"unknown-executable\": executable file not found in $PATH"))
				})
			})

			context("when the executable errors", func() {
				var (
					errorCLI string
					path     string
				)

				it.Before(func() {
					Expect(os.Setenv("PATH", existingPath)).To(Succeed())

					var err error
					errorCLI, err = gexec.Build("github.com/paketo-buildpacks/packit/v2/fakes/some-executable", "-ldflags", "-X main.fail=true")
					Expect(err).NotTo(HaveOccurred())

					path = os.Getenv("PATH")
					Expect(os.Setenv("PATH", filepath.Dir(errorCLI))).To(Succeed())
				})

				it.After(func() {
					Expect(os.Setenv("PATH", path)).To(Succeed())
				})

				it("executes the given arguments against the executable", func() {
					err := executable.Execute(pexec.Execution{
						Args:   []string{"something"},
						Stdout: stdout,
						Stderr: stderr,
					})
					Expect(err).To(MatchError("exit status 1"))
					Expect(stdout).To(ContainSubstring("Error on stdout"))
					Expect(stderr).To(ContainSubstring("Error on stderr"))
				})
			})
		})
	})

	context("Name", func() {
		it("returns the name", func() {
			Expect(executable.Name()).To(Equal(executableName))
		})
	})
}
