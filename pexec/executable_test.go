package pexec_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sclevine/spec"

	"github.com/cloudfoundry/packit/pexec"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/gomega"
)

func testPexec(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		fakeCLI        string
		existingPath   string
		tmpDir         string
		stdout, stderr *bytes.Buffer

		executable pexec.Executable
	)

	it.Before(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "executable")
		Expect(err).NotTo(HaveOccurred())

		tmpDir, err = filepath.EvalSymlinks(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		stdout = bytes.NewBuffer(nil)
		stderr = bytes.NewBuffer(nil)

		executable = pexec.NewExecutable("some-executable")

		fakeCLI, err = gexec.Build("github.com/cloudfoundry/packit/fakes/some-executable")
		Expect(err).NotTo(HaveOccurred())

		existingPath = os.Getenv("PATH")
		os.Setenv("PATH", filepath.Dir(fakeCLI))
	})

	it.After(func() {
		os.Setenv("PATH", existingPath)
		gexec.CleanupBuildArtifacts()
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

		context("failure cases", func() {
			context("when the executable cannot be found on the path", func() {
				it.Before(func() {
					Expect(os.Unsetenv("PATH")).To(Succeed())
				})

				it("returns an error", func() {
					err := executable.Execute(pexec.Execution{})
					Expect(err).To(MatchError("exec: \"some-executable\": executable file not found in $PATH"))
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
					errorCLI, err = gexec.Build("github.com/cloudfoundry/packit/fakes/some-executable", "-ldflags", "-X main.fail=true")
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

}
