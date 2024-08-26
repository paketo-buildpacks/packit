package cargo_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testValidatedReader(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("Read", func() {
		var buffer *bytes.Buffer
		it.Before(func() {
			buffer = bytes.NewBuffer(nil)
		})

		it("reads the contents of the internal reader", func() {
			vr := cargo.NewValidatedReader(strings.NewReader("some-contents"), "sha256:6e32ea34db1b3755d7dec972eb72c705338f0dd8e0be881d966963438fb2e800")

			_, err := io.Copy(buffer, vr)
			Expect(err).NotTo(HaveOccurred())
			Expect(buffer.String()).To(Equal("some-contents"))
		})

		context("when running with a different algorithm", func() {
			it("reads the contents of the internal reader", func() {
				vr := cargo.NewValidatedReader(strings.NewReader("some-contents"), "sha512:b7b2b9e0a4d7f84985a720d1273166bb00132a60ac45388a7d3090a7d4c9692f38d019f807a02750f810f52c623362f977040231c2bbf5947170fe83686cfd9d")

				_, err := io.Copy(buffer, vr)
				Expect(err).NotTo(HaveOccurred())
				Expect(buffer.String()).To(Equal("some-contents"))
			})
		})

		context("when the checksum does not match", func() {
			it("returns an error", func() {
				vr := cargo.NewValidatedReader(strings.NewReader("some-contents"), "sha256:this checksum does not match")

				_, err := io.Copy(buffer, vr)
				Expect(err).To(MatchError("validation error: checksum does not match"))
			})
		})

		context("when the internal reader cannot be read", func() {
			it("returns an error", func() {
				vr := cargo.NewValidatedReader(errorReader{}, "sha256:6e32ea34db1b3755d7dec972eb72c705338f0dd8e0be881d966963438fb2e800")

				_, err := io.Copy(buffer, vr)
				Expect(err).To(MatchError("failed to read"))
			})
		})

		context("failure cases", func() {
			context("there is an unsupported algorithm", func() {
				it("returns an error", func() {
					vr := cargo.NewValidatedReader(strings.NewReader("some-contents"), "magic:6e32ea34db1b3755d7dec972eb72c705338f0dd8e0be881d966963438fb2e800")

					_, err := io.Copy(buffer, vr)
					Expect(err).To(MatchError(`unsupported algorithm "magic": the following algorithms are supported [sha256, sha512]`))
				})
			})
		})
	})

	context("Valid", func() {
		context("when the checksums match", func() {
			it("returns true", func() {
				vr := cargo.NewValidatedReader(strings.NewReader("some-contents"), "sha256:6e32ea34db1b3755d7dec972eb72c705338f0dd8e0be881d966963438fb2e800")

				ok, err := vr.Valid()
				Expect(err).NotTo(HaveOccurred())
				Expect(ok).To(BeTrue())
			})
		})

		context("when the checksums do not match", func() {
			it("returns false", func() {
				vr := cargo.NewValidatedReader(strings.NewReader("some-contents"), "sha256:this checksum does not match")

				ok, err := vr.Valid()
				Expect(err).NotTo(HaveOccurred())
				Expect(ok).To(BeFalse())
			})
		})

		context("failure cases", func() {
			context("when the internal reader cannot be read", func() {
				it("returns an error", func() {
					vr := cargo.NewValidatedReader(errorReader{}, "sha256:6e32ea34db1b3755d7dec972eb72c705338f0dd8e0be881d966963438fb2e800")

					ok, err := vr.Valid()
					Expect(err).To(MatchError("failed to read"))
					Expect(ok).To(BeFalse())
				})
			})
		})
	})
}
