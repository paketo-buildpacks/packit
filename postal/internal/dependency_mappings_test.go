package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/postal/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDependencyMappings(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect      = NewWithT(t).Expect
		path        string
		resolver    internal.DependencyMappingResolver
		bindingPath string
		err         error
	)

	it.Before(func() {
		resolver = internal.NewDependencyMappingResolver()
		bindingPath, err = os.MkdirTemp("", "bindings")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("FindDependencyMapping", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(bindingPath, "some-binding"), 0700)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(bindingPath, "some-binding", "type"), []byte("dependency-mapping"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(bindingPath, "some-binding", "some-sha"), []byte("dependency-mapping-entry.tgz"), 0600)).To(Succeed())

			Expect(os.MkdirAll(filepath.Join(bindingPath, "other-binding"), 0700)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(bindingPath, "other-binding", "type"), []byte("dependency-mapping"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(bindingPath, "other-binding", "other-sha"), []byte("dependency-mapping-entry.tgz"), 0600)).To(Succeed())

			Expect(os.MkdirAll(filepath.Join(bindingPath, "another-binding"), 0700)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(bindingPath, "another-binding", "type"), []byte("another type"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(bindingPath, "another-binding", "some-sha"), []byte("entry.tgz"), 0600)).To(Succeed())
		})

		context("given a set of bindings and a dependency", func() {
			it("finds a matching dependency mappings in the platform bindings if there is one", func() {
				boundDependency, err := resolver.FindDependencyMapping("some-sha", bindingPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(boundDependency).To(Equal("dependency-mapping-entry.tgz"))
			})
		})

		context("given a set of bindings and a dependency", func() {
			it("returns an empty DependencyMapping if there is no match", func() {
				boundDependency, err := resolver.FindDependencyMapping("unmatched-sha", bindingPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(boundDependency).To(Equal(""))
			})
		})
	})

	context("failure cases", func() {
		context("when the binding path is a bad pattern", func() {
			it("errors", func() {
				_, err := resolver.FindDependencyMapping("some-sha", "///")
				Expect(err).To(MatchError(ContainSubstring("failed to list service bindings")))
			})
		})
	})
}
