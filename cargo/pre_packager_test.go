package cargo_test

import (
	"testing"

	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/cargo/fakes"
	"github.com/cloudfoundry/packit/pexec"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPrePackager(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		bash        *fakes.Executable
		prePackager cargo.PrePackager
	)

	it.Before(func() {
		bash = &fakes.Executable{}

		prePackager = cargo.NewPrePackager(bash)
	})

	context("Execute", func() {
		it("executes the pre_package script", func() {
			err := prePackager.Execute("some-script", "some-dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(bash.ExecuteCall.Receives.Execution).To(Equal(pexec.Execution{
				Args: []string{"-c", "some-script"},
				Dir:  "some-dir",
			}))
		})

		it("executes nothing when passed a empty script", func() {
			err := prePackager.Execute("", "some-dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(bash.ExecuteCall.CallCount).To(Equal(0))
		})
	})
}
