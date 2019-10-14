package packit_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/packit"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Executable", func() {
	var (
		executable packit.Executable
		tmpDir     string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "cnb2cf-executable")
		Expect(err).NotTo(HaveOccurred())

		tmpDir, err = filepath.EvalSymlinks(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		logger := lager.NewLogger("cutlass")

		executable = packit.NewExecutable("some-executable", logger)
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	Describe("Execute", func() {
		It("executes the given arguments against the executable", func() {
			stdout, stderr, err := executable.Execute(packit.Execution{
				Args: []string{"something"},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(ContainSubstring("Output on stdout"))
			Expect(stderr).To(ContainSubstring("Output on stderr"))

			Expect(stdout).To(ContainSubstring("Arguments: [some-executable something]"))
		})

		Context("when given a execution directory", func() {
			It("executes within that directory", func() {
				stdout, _, err := executable.Execute(packit.Execution{
					Dir: tmpDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(ContainSubstring(fmt.Sprintf("PWD=%s", tmpDir)))
			})
		})

		Context("when given an execution environment", func() {
			It("executes with that environment", func() {
				stdout, _, err := executable.Execute(packit.Execution{
					Env: []string{"SOME_KEY=some-value"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(ContainSubstring("SOME_KEY=some-value"))
			})
		})

		Context("when given a writer for stdout and stderr", func() {
			It("pipes stdout to that writer", func() {
				stdOutBuffer := bytes.NewBuffer(nil)
				stdErrBuffer := bytes.NewBuffer(nil)

				stdout, stderr, err := executable.Execute(packit.Execution{
					Stdout: stdOutBuffer,
					Stderr: stdErrBuffer,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(stdOutBuffer.String()).To(ContainSubstring("Output on stdout"))
				Expect(stdOutBuffer.String()).To(Equal(stdout))

				Expect(stdErrBuffer.String()).To(ContainSubstring("Output on stderr"))
				Expect(stdErrBuffer.String()).To(Equal(stderr))
			})
		})

		Context("when the executable errors", func() {
			var (
				errorCLI string
				path     string
			)

			BeforeEach(func() {
				os.Setenv("PATH", existingPath)

				var err error
				errorCLI, err = gexec.Build("github.com/cloudfoundry/packit/fakes/some-executable", "-ldflags", "-X main.fail=true")
				Expect(err).NotTo(HaveOccurred())

				path = os.Getenv("PATH")
				os.Setenv("PATH", filepath.Dir(errorCLI))
			})

			AfterEach(func() {
				os.Setenv("PATH", path)
			})

			It("executes the given arguments against the executable", func() {
				stdout, stderr, err := executable.Execute(packit.Execution{
					Args: []string{"something"},
				})
				Expect(err).To(MatchError("exit status 1"))
				Expect(stdout).To(ContainSubstring("Error on stdout"))
				Expect(stderr).To(ContainSubstring("Error on stderr"))
			})
		})
	})
})
