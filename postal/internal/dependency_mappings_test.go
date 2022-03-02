package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/postal/internal"
	"github.com/paketo-buildpacks/packit/v2/postal/internal/fakes"
	"github.com/paketo-buildpacks/packit/v2/servicebindings"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDependencyMappings(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect          = NewWithT(t).Expect
		tmpDir          string
		resolver        internal.DependencyMappingResolver
		bindingResolver *fakes.BindingResolver
		err             error
	)

	it.Before(func() {
		tmpDir, err = os.MkdirTemp("", "dependency-mappings")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(tmpDir, "entry-data"), []byte("\n\tdependency-mapping-entry.tgz\n"), os.ModePerm))

		bindingResolver = &fakes.BindingResolver{}
		resolver = internal.NewDependencyMappingResolver(bindingResolver)
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	context("FindDependencyMapping", func() {
		it.Before(func() {
			bindingResolver.ResolveCall.Returns.BindingSlice = []servicebindings.Binding{
				{
					Name: "some-binding",
					Path: "some-path",
					Type: "dependency-mapping",
					Entries: map[string]*servicebindings.Entry{
						"some-sha": servicebindings.NewEntry(filepath.Join(tmpDir, "entry-data")),
					},
				},
				{
					Name: "other-binding",
					Path: "other-path",
					Type: "dependency-mapping",
					Entries: map[string]*servicebindings.Entry{
						"other-sha": servicebindings.NewEntry("some-entry-path"),
					},
				},
				{
					Name:    "another-binding",
					Path:    "another-path",
					Type:    "another-type",
					Entries: map[string]*servicebindings.Entry{},
				},
			}
		})

		context("given a set of bindings and a dependency", func() {
			it("finds a matching dependency mappings in the platform bindings if there is one", func() {
				boundDependency, err := resolver.FindDependencyMapping("some-sha", "some-platform-dir")
				Expect(err).ToNot(HaveOccurred())
				Expect(bindingResolver.ResolveCall.Receives.Typ).To(Equal("dependency-mapping"))
				Expect(bindingResolver.ResolveCall.Receives.Provider).To(BeEmpty())
				Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal("some-platform-dir"))
				Expect(boundDependency).To(Equal("dependency-mapping-entry.tgz"))
			})
		})

		context("given a set of bindings and a dependency", func() {
			it("returns an empty DependencyMapping if there is no match", func() {
				boundDependency, err := resolver.FindDependencyMapping("unmatched-sha", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(boundDependency).To(Equal(""))
			})
		})
	})
}
