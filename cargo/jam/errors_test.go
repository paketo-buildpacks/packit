package main_test

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testErrors(t *testing.T, context spec.G, it spec.S) {
	var (
		withT      = NewWithT(t)
		Expect     = withT.Expect
		Eventually = withT.Eventually

		buffer *bytes.Buffer
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
	})

	context("failure cases", func() {
		context("when there is no command", func() {
			it("prints an error message", func() {
				command := exec.Command(path)
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1), func() string { return buffer.String() })

				Expect(session.Err).To(gbytes.Say("missing command"))
			})
		})

		context("when the given command is unknown", func() {
			it("prints an error message", func() {
				command := exec.Command(path, "some-unknown-command")
				session, err := gexec.Start(command, buffer, buffer)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1), func() string { return buffer.String() })

				Expect(session.Err).To(gbytes.Say("unknown command: \"some-unknown-command\""))
			})
		})
	})
}
