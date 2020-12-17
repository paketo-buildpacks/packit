package internal_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/paketo-buildpacks/packit/cargo/jam/internal"
	"github.com/paketo-buildpacks/packit/cargo/jam/internal/fakes"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPrePackager(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		bash        *fakes.Executable
		logger      scribe.Logger
		output      *bytes.Buffer
		prePackager internal.PrePackager
	)

	it.Before(func() {
		bash = &fakes.Executable{}
		bash.ExecuteCall.Stub = func(execution pexec.Execution) error {
			if execution.Stdout != nil {
				fmt.Fprint(execution.Stdout, "hello from stdout")
			}

			if execution.Stderr != nil {
				fmt.Fprint(execution.Stderr, "hello from stderr")
			}

			return nil
		}

		output = bytes.NewBuffer(nil)
		logger = scribe.NewLogger(output)
		prePackager = internal.NewPrePackager(bash, logger, output)
	})

	context("Execute", func() {
		it("executes the pre-package script", func() {
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
