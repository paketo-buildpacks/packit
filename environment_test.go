package packit_test

import (
	"testing"

	"github.com/cloudfoundry/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testEnvironment(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		environment = packit.NewEnvironment()
	)

	context("Override", func() {
		it("modifies the environment object with override values", func() {
			environment.Override("SOME_NAME", "some-value")
			Expect(environment).To(Equal(packit.Environment{
				"SOME_NAME.override": "some-value",
			}))
		})

		context("when called multiple times", func() {
			it("overwrites the previous invocation", func() {
				environment.Override("SOME_NAME", "some-value")
				environment.Override("SOME_NAME", "some-other-value")

				Expect(environment).To(Equal(packit.Environment{
					"SOME_NAME.override": "some-other-value",
				}))
			})
		})
	})

	context("Prepend", func() {
		it("modifies the environment object with prepend values", func() {
			environment.Prepend("SOME_NAME", "some-value", "|")
			Expect(environment).To(Equal(packit.Environment{
				"SOME_NAME.prepend": "some-value",
				"SOME_NAME.delim":   "|",
			}))
		})

		context("when called multiple times", func() {
			it("overwrites the previous invocation", func() {
				environment.Prepend("SOME_NAME", "some-value", "|")
				environment.Prepend("SOME_NAME", "other-value", "&")

				Expect(environment).To(Equal(packit.Environment{
					"SOME_NAME.prepend": "other-value",
					"SOME_NAME.delim":   "&",
				}))
			})

			context("when the delimiter is empty", func() {
				it("does not include a .delim variable", func() {
					environment.Prepend("SOME_NAME", "some-value", "|")
					environment.Prepend("SOME_NAME", "other-value", "")

					Expect(environment).To(Equal(packit.Environment{
						"SOME_NAME.prepend": "other-value",
					}))
				})
			})
		})

		context("when the delimiter is empty", func() {
			it("does not include a .delim variable", func() {
				environment.Prepend("SOME_NAME", "some-value", "")
				Expect(environment).To(Equal(packit.Environment{
					"SOME_NAME.prepend": "some-value",
				}))
			})
		})
	})

	context("Append", func() {
		it("modifies an existing environment var with append values", func() {
			environment.Append("SOME_NAME", "some-value", ";")
			Expect(environment).To(Equal(packit.Environment{
				"SOME_NAME.append": "some-value",
				"SOME_NAME.delim":  ";",
			}))
		})

		context("when called multiple times", func() {
			it("overwrites the previous invocation", func() {
				environment.Append("SOME_NAME", "some-value", ";")
				environment.Append("SOME_NAME", "other-value", "&")

				Expect(environment).To(Equal(packit.Environment{
					"SOME_NAME.append": "other-value",
					"SOME_NAME.delim":  "&",
				}))
			})

			context("when the delimiter is empty", func() {
				it("does not include a .delim variable", func() {
					environment.Append("SOME_NAME", "some-value", ";")
					environment.Append("SOME_NAME", "other-value", "")

					Expect(environment).To(Equal(packit.Environment{
						"SOME_NAME.append": "other-value",
					}))
				})
			})
		})

		context("when the delimiter is empty", func() {
			it("does not include a .delim variable", func() {
				environment.Append("SOME_NAME", "some-value", "")
				Expect(environment).To(Equal(packit.Environment{
					"SOME_NAME.append": "some-value",
				}))
			})
		})
	})

	context("Default", func() {
		it("modifies the environment object with default values", func() {
			environment.Default("SOME_NAME", "some-default-value")

			Expect(environment).To(Equal(packit.Environment{
				"SOME_NAME.default": "some-default-value",
			}))
		})

		context("when called multiple times", func() {
			it("overwrites the previous invocation", func() {
				environment.Default("SOME_NAME", "some-default-value")
				environment.Default("SOME_NAME", "other-default-value")

				Expect(environment).To(Equal(packit.Environment{
					"SOME_NAME.default": "other-default-value",
				}))
			})
		})
	})
}
