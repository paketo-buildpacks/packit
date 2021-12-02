package scribe_test

import (
	"bytes"
	"testing"

	"github.com/paketo-buildpacks/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/packit/matchers"
)

func testLogger(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer *bytes.Buffer
		logger scribe.Logger
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		logger = scribe.NewLogger(buffer)
	})

	context("Title", func() {
		it("prints the output with no indentation", func() {
			logger.Title("some-%s", "title")
			Expect(buffer.String()).To(Equal("some-title\n"))
		})
	})

	context("Process", func() {
		it("prints the output with one level of indentation", func() {
			logger.Process("some-%s", "process")
			Expect(buffer.String()).To(Equal("  some-process\n"))
		})
	})

	context("Subprocess", func() {
		it("prints the output with two levels of indentation", func() {
			logger.Subprocess("some-%s", "subprocess")
			Expect(buffer.String()).To(Equal("    some-subprocess\n"))
		})
	})

	context("Action", func() {
		it("prints the output with three levels of indentation", func() {
			logger.Action("some-%s", "action")
			Expect(buffer.String()).To(Equal("      some-action\n"))
		})
	})

	context("Detail", func() {
		it("prints the output with four levels of indentation", func() {
			logger.Detail("some-%s", "detail")
			Expect(buffer.String()).To(Equal("        some-detail\n"))
		})
	})

	context("Subdetail", func() {
		it("prints the output with five levels of indentation", func() {
			logger.Subdetail("some-%s", "subdetail")
			Expect(buffer.String()).To(Equal("          some-subdetail\n"))
		})
	})

	context("Break", func() {
		it("prints an empty line", func() {
			logger.Break()
			Expect(buffer.String()).To(Equal("\n"))
		})
	})

	context("Debug", func() {
		context("when Log Level is not set to DEBUG", func() {
			it("does not print info", func() {
				logger.Title("some-%s", "title")
				logger.Process("some-%s", "process")
				logger.Subprocess("some-%s", "subprocess")
				logger.Action("some-%s", "action")
				logger.Detail("some-%s", "detail")
				logger.Subdetail("some-%s", "subdetail")
				logger.Break()

				logger.Debug.Title("some-debug-%s", "title")
				logger.Debug.Process("some-debug-%s", "process")
				logger.Debug.Subprocess("some-debug-%s", "subprocess")
				logger.Debug.Action("some-debug-%s", "action")
				logger.Debug.Detail("some-debug-%s", "detail")
				logger.Debug.Subdetail("some-debug-%s", "subdetail")
				logger.Debug.Break()
				Expect(buffer.String()).To(ContainLines(
					"some-title",
					"  some-process",
					"    some-subprocess",
					"      some-action",
					"        some-detail",
					"          some-subdetail",
					"",
				))

				Expect(buffer.String()).NotTo(ContainLines(
					"some-debug-title",
					"  some-debug-process",
					"    some-debug-subprocess",
					"      some-debug-action",
					"        some-debug-detail",
					"          some-debug-subdetail",
					"",
				))
			})
		})

		context("when log level set to DEBUG", func() {
			var debugLogger scribe.Logger

			it.Before(func() {
				debugLogger = scribe.NewLogger(buffer).WithLevel("DEBUG")
			})

			it("does print info", func() {
				debugLogger.Title("some-%s", "title")
				debugLogger.Process("some-%s", "process")
				debugLogger.Subprocess("some-%s", "subprocess")
				debugLogger.Action("some-%s", "action")
				debugLogger.Detail("some-%s", "detail")
				debugLogger.Subdetail("some-%s", "subdetail")
				debugLogger.Break()

				debugLogger.Debug.Title("some-debug-%s", "title")
				debugLogger.Debug.Process("some-debug-%s", "process")
				debugLogger.Debug.Subprocess("some-debug-%s", "subprocess")
				debugLogger.Debug.Action("some-debug-%s", "action")
				debugLogger.Debug.Detail("some-debug-%s", "detail")
				debugLogger.Debug.Subdetail("some-debug-%s", "subdetail")
				debugLogger.Debug.Break()
				Expect(buffer.String()).To(ContainLines(
					"some-title",
					"  some-process",
					"    some-subprocess",
					"      some-action",
					"        some-detail",
					"          some-subdetail",
					"",
					"some-debug-title",
					"  some-debug-process",
					"    some-debug-subprocess",
					"      some-debug-action",
					"        some-debug-detail",
					"          some-debug-subdetail",
					"",
				))
			})
		}, spec.Sequential())
	})
}
