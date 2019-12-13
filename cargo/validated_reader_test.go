package cargo_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/cloudfoundry/packit/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testValidatedReader(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		vr     cargo.ValidatedReader
	)

	it.Before(func() {
		vr = cargo.NewValidatedReader(strings.NewReader("some-contents"), "6e32ea34db1b3755d7dec972eb72c705338f0dd8e0be881d966963438fb2e800")
	})

	context("Read", func() {

		it("reads the contents of the internal reader", func() {
			buffer := bytes.NewBuffer(nil)

			_, err := io.Copy(buffer, vr)
			Expect(err).NotTo(HaveOccurred())
			Expect(buffer.String()).To(Equal("some-contents"))
		})

		context("when the checksum does not match", func() {
			it.Before(func() {
				vr = cargo.NewValidatedReader(strings.NewReader("some-contents"), "this checksum does not match")
			})

			it("returns an error", func() {
				buffer := bytes.NewBuffer(nil)

				_, err := io.Copy(buffer, vr)
				Expect(err).To(MatchError("validation error: checksum does not match"))
			})
		})

		context("when the internal reader cannot be read", func() {
			it.Before(func() {
				vr = cargo.NewValidatedReader(errorReader{}, "6e32ea34db1b3755d7dec972eb72c705338f0dd8e0be881d966963438fb2e800")
			})

			it("returns an error", func() {
				buffer := bytes.NewBuffer(nil)

				_, err := io.Copy(buffer, vr)
				Expect(err).To(MatchError("failed to read"))
			})
		})
	})

	context("Valid", func() {
		context("when the checksums match", func() {
			it("returns true", func() {
				ok, err := vr.Valid()
				Expect(err).NotTo(HaveOccurred())
				Expect(ok).To(BeTrue())
			})
		})

		context("when the checksums do not match", func() {
			it.Before(func() {
				vr = cargo.NewValidatedReader(strings.NewReader("some-contents"), "this checksum does not match")
			})

			it("returns false", func() {
				ok, err := vr.Valid()
				Expect(err).NotTo(HaveOccurred())
				Expect(ok).To(BeFalse())
			})
		})

		context("failure cases", func() {
			context("when the internal reader cannot be read", func() {
				it.Before(func() {
					vr = cargo.NewValidatedReader(errorReader{}, "6e32ea34db1b3755d7dec972eb72c705338f0dd8e0be881d966963438fb2e800")
				})

				it("returns an error", func() {
					ok, err := vr.Valid()
					Expect(err).To(MatchError("failed to read"))
					Expect(ok).To(BeFalse())
				})
			})
		})
	})
}
