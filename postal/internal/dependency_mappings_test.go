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
					Path: "other-",
					Type: "dependency-mapping",
					Entries: map[string]*servicebindings.Entry{
						"sha512:other-sha": servicebindings.NewEntry(filepath.Join(tmpDir, "entry-data")),
					},
				},
				{
					Name: "other-binding-with-hyphen",
					Path: "hypen-another-",
					Type: "dependency-mapping",
					Entries: map[string]*servicebindings.Entry{
						"sha-512_other-sha-underscore": servicebindings.NewEntry(filepath.Join(tmpDir, "entry-data")),
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
				boundDependency, err := resolver.FindDependencyMapping("sha256:some-sha", "some-platform-dir")
				Expect(err).ToNot(HaveOccurred())
				Expect(bindingResolver.ResolveCall.Receives.Typ).To(Equal("dependency-mapping"))
				Expect(bindingResolver.ResolveCall.Receives.Provider).To(BeEmpty())
				Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal("some-platform-dir"))
				Expect(boundDependency).To(Equal("dependency-mapping-entry.tgz"))
			})

			context("the binding is of format <algorithm>:<hash>", func() {
				it("finds a matching dependency mappings in the platform bindings if there is one", func() {
					boundDependency, err := resolver.FindDependencyMapping("sha512:other-sha", "some-platform-dir")
					Expect(err).ToNot(HaveOccurred())
					Expect(bindingResolver.ResolveCall.Receives.Typ).To(Equal("dependency-mapping"))
					Expect(bindingResolver.ResolveCall.Receives.Provider).To(BeEmpty())
					Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal("some-platform-dir"))
					Expect(boundDependency).To(Equal("dependency-mapping-entry.tgz"))
				})
			})

			context("the binding is of format <algorithm>_<hash>", func() {
				it("finds a matching dependency mappings in the platform bindings if there is one", func() {
					boundDependency, err := resolver.FindDependencyMapping("sha-512:other-sha-underscore", "some-platform-dir")
					Expect(err).ToNot(HaveOccurred())
					Expect(bindingResolver.ResolveCall.Receives.Typ).To(Equal("dependency-mapping"))
					Expect(bindingResolver.ResolveCall.Receives.Provider).To(BeEmpty())
					Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal("some-platform-dir"))
					Expect(boundDependency).To(Equal("dependency-mapping-entry.tgz"))
				})
			})

			context("the binding does not contain an algorithm", func() {
				it("does not find matching dependency mapping when input isn't of sha256 algorithm", func() {
					boundDependency, err := resolver.FindDependencyMapping("sha512:some-sha", "some-platform-dir")
					Expect(err).ToNot(HaveOccurred())
					Expect(boundDependency).To(Equal(""))
				})
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
