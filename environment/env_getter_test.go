package environment_test

import (
	"os"
	"testing"

	env "github.com/paketo-buildpacks/packit/environment"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testEnvGetter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		envGetter env.EnvGetter
	)

	it.Before(func() {
		envGetter = env.NewEnvGetter()
	})

	context("Lookup", func() {
		context("when env vars are set", func() {
			it.Before(func() {
				os.Setenv("TEST_VARIABLE", "test_value")
				os.Setenv("EMPTY_STRING_VARIABLE", "")
			})
			it.After(func() {
				os.Unsetenv("TEST_VARIABLE")
				os.Unsetenv("EMPTY_STRING_VARIABLE")
			})

			it("returns the value of envirionment variables", func() {
				value, ok := envGetter.Lookup("TEST_VARIABLE")
				Expect(ok).To(BeTrue())
				Expect(value).To(Equal("test_value"))

				value, ok = envGetter.Lookup("EMPTY_STRING_VARIABLE")
				Expect(ok).To(BeTrue())
				Expect(value).To(Equal(""))
			})
		})

		context("when an env var is unset", func() {
			it.Before(func() {
				os.Unsetenv("UNSET_VARIABLE")
			})
			it("returns the empty string and false", func() {
				value, ok := envGetter.Lookup("TEST_VARIABLE")
				Expect(ok).To(BeFalse())
				Expect(value).To(Equal(""))
			})
		})
	})

	context("LookupWithDefault", func() {
		context("when env vars are set", func() {
			it.Before(func() {
				os.Setenv("TEST_VARIABLE", "test_value")
				os.Setenv("EMPTY_STRING_VARIABLE", "")
			})
			it.After(func() {
				os.Unsetenv("TEST_VARIABLE")
				os.Unsetenv("EMPTY_STRING_VARIABLE")
			})

			it("returns the value of envirionment variables", func() {
				value := envGetter.LookupWithDefault("TEST_VARIABLE", "some_default")
				Expect(value).To(Equal("test_value"))

				value = envGetter.LookupWithDefault("EMPTY_STRING_VARIABLE", "some_default")
				Expect(value).To(Equal(""))
			})
		})

		context("when an env var is unset", func() {
			it.Before(func() {
				os.Unsetenv("UNSET_VARIABLE")
			})
			it("returns the default value", func() {
				value := envGetter.LookupWithDefault("TEST_VARIABLE", "some_default")
				Expect(value).To(Equal("some_default"))
			})
		})
	})

	context("GetAsBool", func() {
		context("when an env var is unset", func() {
			it.Before(func() {
				os.Unsetenv("UNSET_VARIABLE")
			})
			it("returns false", func() {
				value := envGetter.GetAsBool("UNSET_VARIABLE")
				Expect(value).To(BeFalse())
			})
		})
		context("when env var is set to a truthy value", func() {
			it.Before(func() {
				os.Setenv("TRUTHY_VAR_0", "")
				os.Setenv("TRUTHY_VAR_1", "1")
				os.Setenv("TRUTHY_VAR_2", "t")
				os.Setenv("TRUTHY_VAR_3", "T")
				os.Setenv("TRUTHY_VAR_4", "TRUE")
				os.Setenv("TRUTHY_VAR_5", "True")
				os.Setenv("TRUTHY_VAR_6", "true")
			})
			it.After(func() {
				os.Unsetenv("TRUTHY_VAR_0")
				os.Unsetenv("TRUTHY_VAR_1")
				os.Unsetenv("TRUTHY_VAR_2")
				os.Unsetenv("TRUTHY_VAR_3")
				os.Unsetenv("TRUTHY_VAR_4")
				os.Unsetenv("TRUTHY_VAR_5")
				os.Unsetenv("TRUTHY_VAR_6")
			})

			it("returns true", func() {
				Expect(envGetter.GetAsBool("TRUTHY_VAR_0")).To(BeTrue())
				Expect(envGetter.GetAsBool("TRUTHY_VAR_1")).To(BeTrue())
				Expect(envGetter.GetAsBool("TRUTHY_VAR_2")).To(BeTrue())
				Expect(envGetter.GetAsBool("TRUTHY_VAR_3")).To(BeTrue())
				Expect(envGetter.GetAsBool("TRUTHY_VAR_4")).To(BeTrue())
				Expect(envGetter.GetAsBool("TRUTHY_VAR_5")).To(BeTrue())
				Expect(envGetter.GetAsBool("TRUTHY_VAR_6")).To(BeTrue())
			})
		})
		context("when env var is set to a falsy value", func() {
			it.Before(func() {
				os.Setenv("FALSY_VAR_0", "0")
				os.Setenv("FALSY_VAR_1", "f")
				os.Setenv("FALSY_VAR_2", "F")
				os.Setenv("FALSY_VAR_3", "FALSE")
				os.Setenv("FALSY_VAR_4", "False")
				os.Setenv("FALSY_VAR_5", "false")
			})
			it.After(func() {
				os.Unsetenv("FALSY_VAR_0")
				os.Unsetenv("FALSY_VAR_1")
				os.Unsetenv("FALSY_VAR_2")
				os.Unsetenv("FALSY_VAR_3")
				os.Unsetenv("FALSY_VAR_4")
				os.Unsetenv("FALSY_VAR_5")
			})

			it("returns true", func() {
				Expect(envGetter.GetAsBool("FALSY_VAR_0")).To(BeFalse())
				Expect(envGetter.GetAsBool("FALSY_VAR_1")).To(BeFalse())
				Expect(envGetter.GetAsBool("FALSY_VAR_2")).To(BeFalse())
				Expect(envGetter.GetAsBool("FALSY_VAR_3")).To(BeFalse())
				Expect(envGetter.GetAsBool("FALSY_VAR_4")).To(BeFalse())
				Expect(envGetter.GetAsBool("FALSY_VAR_5")).To(BeFalse())
			})
		})
	})

	context("GetAsShellWords", func() {
		context("when passed an environment variable whose value is a set of shell words", func() {
			it.Before(func() {
				os.Setenv("SHELL_WORDS_ENV_VAR", "-set /of/shell -w -ords=true --interpolate $OTHER_ENV_VAR")
				os.Setenv("OTHER_ENV_VAR", "other-value")
			})

			it.After(func() {
				os.Unsetenv("SHELL_WORDS_ENV_VAR")
				os.Unsetenv("OTHER_ENV_VAR")
			})

			it("returns the words as string slice elements and interpolates values of env vars", func() {
				words, err := envGetter.GetAsShellWords("SHELL_WORDS_ENV_VAR")
				Expect(err).NotTo(HaveOccurred())

				Expect(words).To(Equal([]string{
					"-set",
					"/of/shell",
					"-w",
					"-ords=true",
					"--interpolate",
					"other-value",
				}))
			})
		})

		context("failure cases", func() {
			context("when env var isn't a well-formed set of flags", func() {
				it.Before(func() {
					os.Setenv("MALFORMED_ENV_VAR", "\"")
				})

				it.After(func() {
					os.Unsetenv("MALFORMED_ENV_VAR")
				})
				it("returns an error", func() {
					_, err := envGetter.GetAsShellWords("MALFORMED_ENV_VAR")
					Expect(err).To(MatchError(ContainSubstring("couldn't parse value of 'MALFORMED_ENV_VAR' as shell words: invalid command line string")))
				})
			})
		})
	})
}
