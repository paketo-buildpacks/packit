package parsers_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/parsers"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testProjectPathParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir        string
		projectDir        string
		projectPathParser parsers.ProjectPathParser
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())
		projectDir = filepath.Join(workingDir, "custom", "path")
		err = os.MkdirAll(projectDir, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		projectPathParser = parsers.NewProjectPathParser()
		os.Setenv("BP_SOME_PROJECT_PATH", "custom/path")
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
		os.Unsetenv("BP_SOME_PROJECT_PATH")
	})

	context("Get", func() {
		it("returns the set project path", func() {
			result, err := projectPathParser.Get(workingDir, "BP_SOME_PROJECT_PATH")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(filepath.Join("custom", "path")))
		})
	})

	context("failure cases", func() {
		context("when the project path subdirectory isn't accessible", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := projectPathParser.Get(workingDir, "BP_SOME_PROJECT_PATH")
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the project path subdirectory does not exist", func() {
			it.Before(func() {
				os.Setenv("BP_SOME_PROJECT_PATH", "some-garbage")
			})

			it.After(func() {
				os.Unsetenv("BP_SOME_PROJECT_PATH")
			})

			it("returns an error", func() {
				_, err := projectPathParser.Get(workingDir, "BP_SOME_PROJECT_PATH")
				Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("expected value derived from BP_SOME_PROJECT_PATH [%s] to be an existing directory", "some-garbage"))))
			})
		})

	})
}
