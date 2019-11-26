package scribe_test

import (
	"testing"

	"github.com/cloudfoundry/packit/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testColor(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	it("returns a function that wraps a string in color codes", func() {
		redFgColor := scribe.NewColor(false, 1, -1)
		Expect(redFgColor("some-text")).To(Equal("\x1b[0;38;5;1msome-text\x1b[0m"))

		blueBgColor := scribe.NewColor(false, -1, 4)
		Expect(blueBgColor("some-text")).To(Equal("\x1b[0;48;5;4msome-text\x1b[0m"))

		magentaBoldFgColor := scribe.NewColor(true, 5, -1)
		Expect(magentaBoldFgColor("some-text")).To(Equal("\x1b[1;38;5;5msome-text\x1b[0m"))

		mixedFgBgColor := scribe.NewColor(false, 3, 244)
		Expect(mixedFgBgColor("some-text")).To(Equal("\x1b[0;38;5;3;48;5;244msome-text\x1b[0m"))
	})

	context("BlackColor", func() {
		it("returns a function that wraps a string in black color codes", func() {
			Expect(scribe.BlackColor("some-text")).To(Equal("\x1b[0;38;5;0msome-text\x1b[0m"))
		})
	})

	context("RedColor", func() {
		it("returns a function that wraps a string in red color codes", func() {
			Expect(scribe.RedColor("some-text")).To(Equal("\x1b[0;38;5;1msome-text\x1b[0m"))
		})
	})

	context("GreenColor", func() {
		it("returns a function that wraps a string in green color codes", func() {
			Expect(scribe.GreenColor("some-text")).To(Equal("\x1b[0;38;5;2msome-text\x1b[0m"))
		})
	})

	context("YellowColor", func() {
		it("returns a function that wraps a string in yellow color codes", func() {
			Expect(scribe.YellowColor("some-text")).To(Equal("\x1b[0;38;5;3msome-text\x1b[0m"))
		})
	})

	context("BlueColor", func() {
		it("returns a function that wraps a string in blue color codes", func() {
			Expect(scribe.BlueColor("some-text")).To(Equal("\x1b[0;38;5;4msome-text\x1b[0m"))
		})
	})

	context("MagentaColor", func() {
		it("returns a function that wraps a string in magenta color codes", func() {
			Expect(scribe.MagentaColor("some-text")).To(Equal("\x1b[0;38;5;5msome-text\x1b[0m"))
		})
	})

	context("CyanColor", func() {
		it("returns a function that wraps a string in cyan color codes", func() {
			Expect(scribe.CyanColor("some-text")).To(Equal("\x1b[0;38;5;6msome-text\x1b[0m"))
		})
	})

	context("WhiteColor", func() {
		it("returns a function that wraps a string in white color codes", func() {
			Expect(scribe.WhiteColor("some-text")).To(Equal("\x1b[0;38;5;7msome-text\x1b[0m"))
		})
	})

	context("GrayColor", func() {
		it("returns a function that wraps a string in gray color codes", func() {
			Expect(scribe.GrayColor("some-text")).To(Equal("\x1b[0;38;5;244msome-text\x1b[0m"))
		})
	})
}
