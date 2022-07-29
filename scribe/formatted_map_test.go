package scribe_test

import (
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testFormattedMap(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("String", func() {
		it("returns a formatted string representation of the map", func() {
			Expect(scribe.FormattedMap{
				"third":  3,
				"first":  1,
				"second": 2,
			}.String()).To(Equal("first  -> \"1\"\nsecond -> \"2\"\nthird  -> \"3\""))
		})
	})

	context("NewFormattedMapFromEnvironment", func() {
		context("for all packit env var operations", func() {
			it("prints the env in a well formatted map", func() {
				Expect(scribe.NewFormattedMapFromEnvironment(packit.Environment{
					"OVERRIDE.override": "some-value",
					"DEFAULT.default":   "some-value",
					"PREPEND.prepend":   "some-value",
					"PREPEND.delim":     ":",
					"APPEND.append":     "some-value",
					"APPEND.delim":      ":",
				})).To(Equal(scribe.FormattedMap{
					"OVERRIDE": "some-value",
					"DEFAULT":  "some-value",
					"PREPEND":  "some-value:$PREPEND",
					"APPEND":   "$APPEND:some-value",
				}))
			})
			it("excludes ill-formed env var names", func() {
				Expect(scribe.NewFormattedMapFromEnvironment(packit.Environment{
					"OVERRIDE.override":         "some-value",
					"DEFAULT.default":           "some-value",
					"PREPEND.prepend":           "some-value",
					"PREPEND.delim":             ":",
					"APPEND.append":             "some-value",
					"APPEND.delim":              ":",
					"invalid=OVERRIDE.override": "some-value",
					"invalid=DEFAULT.default":   "some-value",
					"PRE PEND.prepend":          "some-value",
					"PRE PEND.delim":            ":",
					"APP*END.append":            "some-value",
					"APP*END.delim":             ":",
				})).To(Equal(scribe.FormattedMap{
					"OVERRIDE": "some-value",
					"DEFAULT":  "some-value",
					"PREPEND":  "some-value:$PREPEND",
					"APPEND":   "$APPEND:some-value",
				}))
			})
		})
		context("for a standard string map", func() {
			it("prints the env in a well formatted map", func() {
				Expect(scribe.NewFormattedMapFromEnvironment(map[string]string{
					"SOME_ENV_VAR":       "some-value",
					"SOME_OTHER_ENV_VAR": "some-other-value",
				})).To(Equal(scribe.FormattedMap{
					"SOME_ENV_VAR":       "some-value",
					"SOME_OTHER_ENV_VAR": "some-other-value",
				}))
			})
			it("excludes ill-formed env vars", func() {
				Expect(scribe.NewFormattedMapFromEnvironment(map[string]string{
					"invalid=ENV_VAR_NAME": "some-value",
					"SOME_OTHER_ENV_VAR":   "some-other-value",
				})).To(Equal(scribe.FormattedMap{
					"SOME_OTHER_ENV_VAR": "some-other-value",
				}))
			})
		})
	})
}
