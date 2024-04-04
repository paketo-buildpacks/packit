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

func testDependencyMirror(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect          = NewWithT(t).Expect
		tmpDir          string
		resolver        internal.DependencyMirrorResolver
		bindingResolver *fakes.BindingResolver
		err             error
	)

	context("FindDependencyMirror", func() {
		context("via binding", func() {
			it.Before(func() {
				tmpDir, err = os.MkdirTemp("", "dependency-mirror")
				Expect(err).NotTo(HaveOccurred())
				Expect(os.WriteFile(filepath.Join(tmpDir, "default"), []byte("https://mirror.example.org/{originalHost}"), os.ModePerm))
				Expect(os.WriteFile(filepath.Join(tmpDir, "type"), []byte("dependency-mirror"), os.ModePerm))

				bindingResolver = &fakes.BindingResolver{}
				resolver = internal.NewDependencyMirrorResolver(bindingResolver)

				bindingResolver.ResolveCall.Returns.BindingSlice = []servicebindings.Binding{
					{
						Name: "some-binding",
						Path: "some-path",
						Type: "dependency-mirror",
						Entries: map[string]*servicebindings.Entry{
							"default": servicebindings.NewEntry(filepath.Join(tmpDir, "default")),
						},
					},
				}
			})

			it.After(func() {
				Expect(os.RemoveAll(tmpDir)).To(Succeed())
			})

			context("given a default mirror binding", func() {
				it("finds a matching dependency mirror in the platform bindings if there is one", func() {
					boundDependency, err := resolver.FindDependencyMirror("some-uri", "some-platform-dir")
					Expect(err).ToNot(HaveOccurred())
					Expect(bindingResolver.ResolveCall.Receives.Typ).To(Equal("dependency-mirror"))
					Expect(bindingResolver.ResolveCall.Receives.Provider).To(BeEmpty())
					Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal("some-platform-dir"))
					Expect(boundDependency).To(Equal("https://mirror.example.org/some-uri"))
				})
			})

			context("given default mirror and specific hostname bindings", func() {
				it.Before(func() {
					Expect(os.WriteFile(filepath.Join(tmpDir, "github.com"), []byte("https://mirror.example.org/public-github"), os.ModePerm))
					Expect(os.WriteFile(filepath.Join(tmpDir, "nodejs.org"), []byte("https://mirror.example.org/node-dist"), os.ModePerm))

					bindingResolver.ResolveCall.Returns.BindingSlice[0].Entries = map[string]*servicebindings.Entry{
						"default":    servicebindings.NewEntry(filepath.Join(tmpDir, "default")),
						"github.com": servicebindings.NewEntry(filepath.Join(tmpDir, "github.com")),
						"nodejs.org": servicebindings.NewEntry(filepath.Join(tmpDir, "nodejs.org")),
					}
				})

				it("finds the default mirror when given uri does not match a specific hostname", func() {
					boundDependency, err := resolver.FindDependencyMirror("some-uri", "some-platform-dir")
					Expect(err).ToNot(HaveOccurred())
					Expect(bindingResolver.ResolveCall.Receives.Typ).To(Equal("dependency-mirror"))
					Expect(bindingResolver.ResolveCall.Receives.Provider).To(BeEmpty())
					Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal("some-platform-dir"))
					Expect(boundDependency).To(Equal("https://mirror.example.org/some-uri"))
				})

				it("finds the mirror matching the specific hostname in the given uri", func() {
					boundDependency, err := resolver.FindDependencyMirror("some-github.com-uri", "some-platform-dir")
					Expect(err).ToNot(HaveOccurred())
					Expect(bindingResolver.ResolveCall.Receives.Typ).To(Equal("dependency-mirror"))
					Expect(bindingResolver.ResolveCall.Receives.Provider).To(BeEmpty())
					Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal("some-platform-dir"))
					Expect(boundDependency).To(Equal("https://mirror.example.org/public-github"))
				})
			})

			context("given a specific hostname binding and no default mirror binding", func() {
				it.Before(func() {
					Expect(os.Remove(filepath.Join(tmpDir, "default")))
					Expect(os.WriteFile(filepath.Join(tmpDir, "github.com"), []byte("https://mirror.example.org/public-github"), os.ModePerm))
					Expect(os.WriteFile(filepath.Join(tmpDir, "nodejs.org"), []byte("https://mirror.example.org/node-dist"), os.ModePerm))

					bindingResolver.ResolveCall.Returns.BindingSlice[0].Entries = map[string]*servicebindings.Entry{
						"github.com": servicebindings.NewEntry(filepath.Join(tmpDir, "github.com")),
						"nodejs.org": servicebindings.NewEntry(filepath.Join(tmpDir, "nodejs.org")),
					}
				})

				it("return empty string for non specific hostnames", func() {
					boundDependency, err := resolver.FindDependencyMirror("some-uri", "some-platform-dir")
					Expect(err).ToNot(HaveOccurred())
					Expect(bindingResolver.ResolveCall.Receives.Typ).To(Equal("dependency-mirror"))
					Expect(bindingResolver.ResolveCall.Receives.Provider).To(BeEmpty())
					Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal("some-platform-dir"))
					Expect(boundDependency).To(Equal(""))
				})

				it("finds the mirror matching the specific hostname in the given uri", func() {
					boundDependency, err := resolver.FindDependencyMirror("some-nodejs.org-uri", "some-platform-dir")
					Expect(err).ToNot(HaveOccurred())
					Expect(bindingResolver.ResolveCall.Receives.Typ).To(Equal("dependency-mirror"))
					Expect(bindingResolver.ResolveCall.Receives.Provider).To(BeEmpty())
					Expect(bindingResolver.ResolveCall.Receives.PlatformDir).To(Equal("some-platform-dir"))
					Expect(boundDependency).To(Equal("https://mirror.example.org/node-dist"))
				})
			})
		})

		context("via environment variables", func() {
			it.Before(func() {
				Expect(os.Setenv("BP_DEPENDENCY_MIRROR", "https://mirror.example.org/{originalHost}"))

				bindingResolver = &fakes.BindingResolver{}
				resolver = internal.NewDependencyMirrorResolver(bindingResolver)
			})

			it.After(func() {
				Expect(os.Unsetenv("BP_DEPENDENCY_MIRROR"))
			})

			context("given the default mirror environment variable is set", func() {
				it("finds the matching dependency mirror", func() {
					boundDependency, err := resolver.FindDependencyMirror("some-uri", "some-platform-dir")
					Expect(err).ToNot(HaveOccurred())
					Expect(boundDependency).To(Equal("https://mirror.example.org/some-uri"))
				})
			})

			context("given environment variables for a default mirror and specific hostname mirrors", func() {
				it.Before(func() {
					Expect(os.Setenv("BP_DEPENDENCY_MIRROR_GITHUB_COM", "https://mirror.example.org/public-github"))
					Expect(os.Setenv("BP_DEPENDENCY_MIRROR_TESTING_123__ABC", "https://mirror.example.org/testing"))
				})

				it.After(func() {
					Expect(os.Unsetenv("BP_DEPENDENCY_MIRROR_GITHUB_COM"))
					Expect(os.Unsetenv("BP_DEPENDENCY_MIRROR_TESTING_123__ABC"))
				})

				it("finds the default mirror when given uri does not match a specific hostname", func() {
					boundDependency, err := resolver.FindDependencyMirror("some-uri", "some-platform-dir")
					Expect(err).ToNot(HaveOccurred())
					Expect(boundDependency).To(Equal("https://mirror.example.org/some-uri"))
				})

				it("finds the mirror matching the specific hostname in the given uri", func() {
					boundDependency, err := resolver.FindDependencyMirror("some-github.com-uri", "some-platform-dir")
					Expect(err).ToNot(HaveOccurred())
					Expect(boundDependency).To(Equal("https://mirror.example.org/public-github"))
				})

				it("properly decodes the hostname", func() {
					boundDependency, err := resolver.FindDependencyMirror("testing.123-abc-uri", "some-platform-dir")
					Expect(err).ToNot(HaveOccurred())
					Expect(boundDependency).To(Equal("https://mirror.example.org/testing"))
				})
			})

			context("given environment variables for a specific hostname and none for a default mirror", func() {
				it.Before(func() {
					Expect(os.Unsetenv("BP_DEPENDENCY_MIRROR"))
					Expect(os.Setenv("BP_DEPENDENCY_MIRROR_GITHUB_COM", "https://mirror.example.org/public-github"))
				})

				it.After(func() {
					Expect(os.Unsetenv("BP_DEPENDENCY_MIRROR_GITHUB_COM"))
				})

				it("return empty string for non specific hostnames", func() {
					boundDependency, err := resolver.FindDependencyMirror("some-uri", "some-platform-dir")
					Expect(err).ToNot(HaveOccurred())
					Expect(boundDependency).To(Equal(""))
				})

				it("finds the mirror matching the specific hostname in the given uri", func() {
					boundDependency, err := resolver.FindDependencyMirror("some-github.com-uri", "some-platform-dir")
					Expect(err).ToNot(HaveOccurred())
					Expect(boundDependency).To(Equal("https://mirror.example.org/public-github"))
				})
			})
		})

		context("when mirror is provided by both bindings and environment variables", func() {
			it.Before(func() {
				tmpDir, err = os.MkdirTemp("", "dependency-mirror")
				Expect(err).NotTo(HaveOccurred())
				Expect(os.WriteFile(filepath.Join(tmpDir, "default"), []byte("https://mirror.example.org/{originalHost}"), os.ModePerm))
				Expect(os.WriteFile(filepath.Join(tmpDir, "type"), []byte("dependency-mirror"), os.ModePerm))

				Expect(os.Setenv("BP_DEPENDENCY_MIRROR", "https://mirror.other-example.org/{originalHost}"))
			})

			it.After(func() {
				Expect(os.RemoveAll(tmpDir)).To(Succeed())
				Expect(os.Unsetenv("BP_DEPENDENCY_MIRROR"))
			})

			it("defaults to environment variable and ignores binding", func() {
				boundDependency, err := resolver.FindDependencyMirror("some-uri", "some-platform-dir")
				Expect(err).ToNot(HaveOccurred())
				Expect(boundDependency).NotTo(Equal("https://mirror.example.org/some-uri"))
				Expect(boundDependency).To(Equal("https://mirror.other-example.org/some-uri"))

			})
		})

		context("failure cases", func() {
			context("when more than one dependency mirror binding exists", func() {
				it.Before(func() {
					tmpDir, err = os.MkdirTemp("", "dependency-mirror")
					Expect(err).NotTo(HaveOccurred())
					Expect(os.WriteFile(filepath.Join(tmpDir, "default"), []byte("https://mirror.example.org/{originalHost}"), os.ModePerm))
					Expect(os.WriteFile(filepath.Join(tmpDir, "github.com"), []byte("https://mirror.example.org/public-github"), os.ModePerm))
					Expect(os.WriteFile(filepath.Join(tmpDir, "type"), []byte("dependency-mirror"), os.ModePerm))

					bindingResolver = &fakes.BindingResolver{}
					resolver = internal.NewDependencyMirrorResolver(bindingResolver)
				})

				it.After(func() {
					Expect(os.RemoveAll(tmpDir)).To(Succeed())
				})

				it("returns an error", func() {
					bindingResolver.ResolveCall.Returns.BindingSlice = []servicebindings.Binding{
						{
							Name: "some-binding",
							Path: "some-path",
							Type: "dependency-mirror",
							Entries: map[string]*servicebindings.Entry{
								"default": servicebindings.NewEntry(filepath.Join(tmpDir, "default")),
							},
						},
						{
							Name: "some-other-binding",
							Path: "some-other-path",
							Type: "dependency-mirror",
							Entries: map[string]*servicebindings.Entry{
								"github.com": servicebindings.NewEntry(filepath.Join(tmpDir, "github.com")),
							},
						},
					}

					_, err = resolver.FindDependencyMirror("some-uri", "some-platform-dir")
					Expect(err).To(MatchError(ContainSubstring("cannot have multiple bindings of type 'dependency-mirror'")))
				})
			})

			context("when mirror contains invalid scheme", func() {
				it.Before(func() {
					Expect(os.Setenv("BP_DEPENDENCY_MIRROR", "http://mirror.example.org/{originalHost}"))

					bindingResolver = &fakes.BindingResolver{}
					resolver = internal.NewDependencyMirrorResolver(bindingResolver)
				})

				it.After(func() {
					Expect(os.Unsetenv("BP_DEPENDENCY_MIRROR"))
				})

				it("returns an error", func() {
					_, err := resolver.FindDependencyMirror("some-uri", "some-platform-dir")
					Expect(err).To(MatchError(ContainSubstring("invalid mirror scheme")))
				})
			})
		})
	})
}
