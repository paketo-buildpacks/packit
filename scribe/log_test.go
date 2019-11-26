package scribe_test

import (
	"bytes"
	"testing"

	"github.com/cloudfoundry/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLog(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("Log", func() {
		var (
			buffer *bytes.Buffer
			log    scribe.Log
		)

		it.Before(func() {
			buffer = bytes.NewBuffer(nil)
			log = scribe.NewLog(buffer)
		})

		context("Write", func() {
			it("prints to the writer", func() {
				_, err := log.Write([]byte("some-text"))
				Expect(err).NotTo(HaveOccurred())
				Expect(buffer.String()).To(Equal("some-text"))
			})

			context("when the log has a color", func() {
				it.Before(func() {
					log = scribe.NewLog(buffer, scribe.WithColor(scribe.BlueColor))
				})

				it("prints to the writer with the correct color codes", func() {
					_, err := log.Write([]byte("some-text"))
					Expect(err).NotTo(HaveOccurred())
					Expect(buffer.String()).To(Equal("\x1b[0;38;5;4msome-text\x1b[0m"))
				})
			})

			context("when the log has an indent", func() {
				it.Before(func() {
					log = scribe.NewLog(buffer, scribe.WithIndent(2))
				})

				it("prints to the writer with the correct indentation", func() {
					_, err := log.Write([]byte("some-text\nother-text"))
					Expect(err).NotTo(HaveOccurred())
					Expect(buffer.String()).To(Equal("    some-text\n    other-text"))
				})
			})

			context("when the log has a return prefix", func() {
				it.Before(func() {
					log = scribe.NewLog(buffer, scribe.WithColor(scribe.RedColor), scribe.WithIndent(2))
				})

				it("prints to the writer with the correct indentation", func() {
					_, err := log.Write([]byte("\rsome-text"))
					Expect(err).NotTo(HaveOccurred())
					Expect(buffer.String()).To(Equal("\r\x1b[0;38;5;1m    some-text\x1b[0m"))
				})
			})

			context("when the log has a newline suffix", func() {
				it.Before(func() {
					log = scribe.NewLog(buffer, scribe.WithColor(scribe.RedColor), scribe.WithIndent(2))
				})

				it("prints to the writer with the correct indentation", func() {
					_, err := log.Write([]byte("some-text\n"))
					Expect(err).NotTo(HaveOccurred())
					Expect(buffer.String()).To(Equal("\x1b[0;38;5;1m    some-text\x1b[0m\n"))
				})
			})

			context("when the input has a percent symbol", func() {
				it.Before(func() {
					log = scribe.NewLog(buffer, scribe.WithColor(scribe.MagentaColor))
				})

				it("prints to the writer with the correct indentation", func() {
					_, err := log.Write([]byte("some-%"))
					Expect(err).NotTo(HaveOccurred())
					Expect(buffer.String()).To(Equal("\x1b[0;38;5;5msome-%\x1b[0m"))
				})
			})
		})

		context("Print", func() {
			it("prints to the writer", func() {
				log.Print("some-text")
				Expect(buffer.String()).To(Equal("some-text"))
			})

			context("when the log has a color", func() {
				it.Before(func() {
					log = scribe.NewLog(buffer, scribe.WithColor(scribe.BlueColor))
				})

				it("prints to the writer with the correct color codes", func() {
					log.Print("some-text")
					Expect(buffer.String()).To(Equal("\x1b[0;38;5;4msome-text\x1b[0m"))
				})
			})

			context("when the log has an indent", func() {
				it.Before(func() {
					log = scribe.NewLog(buffer, scribe.WithIndent(2))
				})

				it("prints to the writer with the correct indentation", func() {
					log.Print("some-text")
					Expect(buffer.String()).To(Equal("    some-text"))
				})
			})

			context("when the log has a return character", func() {
				it.Before(func() {
					log = scribe.NewLog(buffer, scribe.WithColor(scribe.RedColor), scribe.WithIndent(2))
				})

				it("prints to the writer with the correct indentation", func() {
					log.Print("\rsome-text")
					Expect(buffer.String()).To(Equal("\r\x1b[0;38;5;1m    some-text\x1b[0m"))
				})
			})
		})

		context("Println", func() {
			it("prints to the writer", func() {
				log.Println("some-text")
				Expect(buffer.String()).To(Equal("some-text\n"))
			})

			context("when the log has a color", func() {
				it.Before(func() {
					log = scribe.NewLog(buffer, scribe.WithColor(scribe.BlueColor))
				})

				it("prints to the writer with the correct color codes", func() {
					log.Println("some-text")
					Expect(buffer.String()).To(Equal("\x1b[0;38;5;4msome-text\x1b[0m\n"))
				})
			})

			context("when the log has an indent", func() {
				it.Before(func() {
					log = scribe.NewLog(buffer, scribe.WithIndent(2))
				})

				it("prints to the writer with the correct indentation", func() {
					log.Println("some-text")
					Expect(buffer.String()).To(Equal("    some-text\n"))
				})
			})

			context("when the log has a return character", func() {
				it.Before(func() {
					log = scribe.NewLog(buffer, scribe.WithColor(scribe.RedColor), scribe.WithIndent(2))
				})

				it("prints to the writer with the correct indentation", func() {
					log.Println("\rsome-text")
					Expect(buffer.String()).To(Equal("\r\x1b[0;38;5;1m    some-text\x1b[0m\n"))
				})
			})
		})

		context("Printf", func() {
			it("prints to the writer", func() {
				log.Printf("some-%s", "text")
				Expect(buffer.String()).To(Equal("some-text"))
			})

			context("when the log has a color", func() {
				it.Before(func() {
					log = scribe.NewLog(buffer, scribe.WithColor(scribe.BlueColor))
				})

				it("prints to the writer with the correct color codes", func() {
					log.Printf("some-%s", "text")
					Expect(buffer.String()).To(Equal("\x1b[0;38;5;4msome-text\x1b[0m"))
				})
			})

			context("when the log has an indent", func() {
				it.Before(func() {
					log = scribe.NewLog(buffer, scribe.WithIndent(2))
				})

				it("prints to the writer with the correct indentation", func() {
					log.Printf("some-%s", "text")
					Expect(buffer.String()).To(Equal("    some-text"))
				})
			})

			context("when the log has a return or newline prefix/suffix", func() {
				it.Before(func() {
					log = scribe.NewLog(buffer, scribe.WithColor(scribe.RedColor), scribe.WithIndent(2))
				})

				it("prints to the writer with the correct indentation", func() {
					log.Printf("\rsome-%s\n", "text")
					Expect(buffer.String()).To(Equal("\r\x1b[0;38;5;1m    some-text\x1b[0m\n"))
				})
			})
		})

		context("Break", func() {
			it("prints a newline", func() {
				log.Break()
				Expect(buffer.String()).To(Equal("\n"))
			})
		})
	})
}
