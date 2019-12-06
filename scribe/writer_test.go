package scribe_test

import (
	"bytes"
	"testing"

	"github.com/cloudfoundry/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testWriter(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("Writer", func() {
		var (
			buffer *bytes.Buffer
			writer scribe.Writer
		)

		it.Before(func() {
			buffer = bytes.NewBuffer(nil)
			writer = scribe.NewWriter(buffer)
		})

		context("Write", func() {
			it("prints to the writer", func() {
				_, err := writer.Write([]byte("some-text"))
				Expect(err).NotTo(HaveOccurred())
				Expect(buffer.String()).To(Equal("some-text"))
			})

			context("when the writer has a color", func() {
				it.Before(func() {
					writer = scribe.NewWriter(buffer, scribe.WithColor(scribe.BlueColor))
				})

				it("prints to the writer with the correct color codes", func() {
					_, err := writer.Write([]byte("some-text"))
					Expect(err).NotTo(HaveOccurred())
					Expect(buffer.String()).To(Equal("\x1b[0;38;5;4msome-text\x1b[0m"))
				})
			})

			context("when the writer has an indent", func() {
				it.Before(func() {
					writer = scribe.NewWriter(buffer, scribe.WithIndent(2))
				})

				it("prints to the writer with the correct indentation", func() {
					_, err := writer.Write([]byte("some-text\nother-text"))
					Expect(err).NotTo(HaveOccurred())
					Expect(buffer.String()).To(Equal("    some-text\n    other-text"))
				})
			})

			context("when the writer has a return prefix", func() {
				it.Before(func() {
					writer = scribe.NewWriter(buffer, scribe.WithColor(scribe.RedColor), scribe.WithIndent(2))
				})

				it("prints to the writer with the correct indentation", func() {
					_, err := writer.Write([]byte("\rsome-text"))
					Expect(err).NotTo(HaveOccurred())
					Expect(buffer.String()).To(Equal("\r\x1b[0;38;5;1m    some-text\x1b[0m"))
				})
			})

			context("when the writer has a newline suffix", func() {
				it.Before(func() {
					writer = scribe.NewWriter(buffer, scribe.WithColor(scribe.RedColor), scribe.WithIndent(2))
				})

				it("prints to the writer with the correct indentation", func() {
					_, err := writer.Write([]byte("some-text\n"))
					Expect(err).NotTo(HaveOccurred())
					Expect(buffer.String()).To(Equal("\x1b[0;38;5;1m    some-text\x1b[0m\n"))
				})
			})

			context("when the input has a percent symbol", func() {
				it.Before(func() {
					writer = scribe.NewWriter(buffer, scribe.WithColor(scribe.MagentaColor))
				})

				it("prints to the writer with the correct indentation", func() {
					_, err := writer.Write([]byte("some-%"))
					Expect(err).NotTo(HaveOccurred())
					Expect(buffer.String()).To(Equal("\x1b[0;38;5;5msome-%\x1b[0m"))
				})
			})
		})
	})
}
