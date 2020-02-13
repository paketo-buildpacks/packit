package cargo_test

import (
	"bytes"
	"testing"

	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/cargo/fakes"
	"github.com/cloudfoundry/packit/pexec"
	"github.com/cloudfoundry/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPrePackager(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		bash        *fakes.Executable
		logger      scribe.Logger
		output      *bytes.Buffer
		prePackager cargo.PrePackager
	)

	it.Before(func() {
		bash = &fakes.Executable{}
		bash.ExecuteCall.Stub = func(execution pexec.Execution) error {
			if execution.Stdout != nil {
				execution.Stdout.Write([]byte("hello from stdout"))
			}

			if execution.Stderr != nil {
				execution.Stderr.Write([]byte("hello from stderr"))
			}

			return nil
		}

		output = bytes.NewBuffer(nil)
		logger = scribe.NewLogger(output)
		prePackager = cargo.NewPrePackager(bash, logger, output)
	})

	context("Execute", func() {
		it("executes the pre_package script", func() {
			err := prePackager.Execute("some-script", "some-dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(bash.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"-c", "some-script"}))
			Expect(bash.ExecuteCall.Receives.Execution.Dir).To(Equal("some-dir"))

			Expect(output.String()).To(ContainSubstring("Executing pre-packaging script: some-script"))
			Expect(output.String()).To(ContainSubstring("hello from stdout"))
			Expect(output.String()).To(ContainSubstring("hello from stderr"))
		})

		it("executes nothing when passed a empty script", func() {
			err := prePackager.Execute("", "some-dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(bash.ExecuteCall.CallCount).To(Equal(0))
			Expect(output.String()).To(BeEmpty())
		})
	})
}
