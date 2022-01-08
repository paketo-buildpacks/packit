package servicebindings_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/servicebindings"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testResolver(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("binding root precedence", func() {
		var (
			bindingRootK8s string
			bindingRootCNB string
			platformDir    string
		)

		it.Before(func() {
			var err error

			bindingRootK8s, err = os.MkdirTemp("", "bindings-k8s")
			Expect(err).NotTo(HaveOccurred())

			err = os.MkdirAll(filepath.Join(bindingRootK8s, "some-binding"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRootK8s, "some-binding", "type"), []byte("some-type"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			bindingRootCNB, err = os.MkdirTemp("", "bindings-cnb")
			Expect(err).NotTo(HaveOccurred())

			err = os.MkdirAll(filepath.Join(bindingRootCNB, "some-binding"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRootCNB, "some-binding", "type"), []byte("some-type"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			platformDir, err = os.MkdirTemp("", "bindings-platform")
			Expect(err).NotTo(HaveOccurred())

			err = os.MkdirAll(filepath.Join(platformDir, "bindings", "some-binding"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(platformDir, "bindings", "some-binding", "type"), []byte("some-type"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		context("SERVICE_BINDING_ROOT env var is set", func() {
			it.Before(func() {
				Expect(os.Setenv("SERVICE_BINDING_ROOT", bindingRootK8s)).To(Succeed())
			})

			context("CNB_BINDINGS env var is set", func() {
				it.Before(func() {
					Expect(os.Setenv("CNB_BINDINGS", bindingRootCNB)).To(Succeed())
				})

				it("resolves bindings from SERVICE_BINDING_ROOT", func() {
					resolver := servicebindings.NewResolver()

					bindings, err := resolver.Resolve("some-type", "", platformDir)
					Expect(err).NotTo(HaveOccurred())
					Expect(bindings).To(ConsistOf(
						servicebindings.Binding{
							Name:    "some-binding",
							Path:    filepath.Join(bindingRootK8s, "some-binding"),
							Type:    "some-type",
							Entries: map[string]*servicebindings.Entry{},
						},
					))
				})
			})

			context("CNB_BINDINGS env var is not set", func() {
				it.Before(func() {
					Expect(os.Unsetenv("CNB_BINDINGS")).To(Succeed())
				})

				it("resolves bindings from SERVICE_BINDING_ROOT", func() {
					resolver := servicebindings.NewResolver()

					bindings, err := resolver.Resolve("some-type", "", platformDir)
					Expect(err).NotTo(HaveOccurred())
					Expect(bindings).To(ConsistOf(
						servicebindings.Binding{
							Name:    "some-binding",
							Path:    filepath.Join(bindingRootK8s, "some-binding"),
							Type:    "some-type",
							Entries: map[string]*servicebindings.Entry{},
						},
					))
				})
			})
		})

		context("SERVICE_BINDING_ROOT env var is not set", func() {
			it.Before(func() {
				Expect(os.Unsetenv("SERVICE_BINDING_ROOT")).To(Succeed())
			})

			context("CNB_BINDINGS env var is set", func() {
				it.Before(func() {
					Expect(os.Setenv("CNB_BINDINGS", bindingRootCNB)).To(Succeed())
				})

				it("resolves bindings from CNB_BINDINGS", func() {
					resolver := servicebindings.NewResolver()

					bindings, err := resolver.Resolve("some-type", "", platformDir)
					Expect(err).NotTo(HaveOccurred())
					Expect(bindings).To(ConsistOf(
						servicebindings.Binding{
							Name:    "some-binding",
							Path:    filepath.Join(bindingRootCNB, "some-binding"),
							Type:    "some-type",
							Entries: map[string]*servicebindings.Entry{},
						},
					))
				})
			})

			context("CNB_BINDINGS env var is not set", func() {
				it.Before(func() {
					Expect(os.Unsetenv("CNB_BINDINGS")).To(Succeed())
				})

				it("resolves bindings from platform dir", func() {
					resolver := servicebindings.NewResolver()

					bindings, err := resolver.Resolve("some-type", "", platformDir)
					Expect(err).NotTo(HaveOccurred())
					Expect(bindings).To(ConsistOf(
						servicebindings.Binding{
							Name:    "some-binding",
							Path:    filepath.Join(platformDir, "bindings", "some-binding"),
							Type:    "some-type",
							Entries: map[string]*servicebindings.Entry{},
						},
					))
				})
			})
		})
	})

	context("resolving bindings", func() {
		var bindingRoot string
		var resolver *servicebindings.Resolver

		it.Before(func() {
			var err error
			bindingRoot, err = os.MkdirTemp("", "bindings")
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Setenv("SERVICE_BINDING_ROOT", bindingRoot)).To(Succeed())

			resolver = servicebindings.NewResolver()

			err = os.MkdirAll(filepath.Join(bindingRoot, "binding-1A"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRoot, "binding-1A", "type"), []byte("type-1"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRoot, "binding-1A", "provider"), []byte("provider-1A"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRoot, "binding-1A", "username"), nil, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRoot, "binding-1A", "password"), nil, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.MkdirAll(filepath.Join(bindingRoot, "binding-1B"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRoot, "binding-1B", "type"), []byte("type-1"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRoot, "binding-1B", "provider"), []byte("provider-1B"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRoot, "binding-1B", "username"), nil, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRoot, "binding-1B", "password"), nil, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.MkdirAll(filepath.Join(bindingRoot, "binding-2"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRoot, "binding-2", "type"), []byte("type-2"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRoot, "binding-2", "provider"), []byte("provider-2"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRoot, "binding-2", "username"), nil, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRoot, "binding-2", "password"), nil, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.MkdirAll(filepath.Join(bindingRoot, "binding-3"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRoot, "binding-3", "type"), []byte("\n type-3\n"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRoot, "binding-3", "provider"), []byte("\tprovider-3\n"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.WriteFile(filepath.Join(bindingRoot, "binding-3", "value"), nil, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

		})

		it.After(func() {
			Expect(os.RemoveAll(bindingRoot)).To(Succeed())
			Expect(os.Unsetenv("SERVICE_BINDING_ROOT")).To(Succeed())
		})

		context("Resolve", func() {
			it("resolves by type only (case-insensitive)", func() {
				bindings, err := resolver.Resolve("TyPe-1", "", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(bindings).To(ConsistOf(
					servicebindings.Binding{
						Name:     "binding-1A",
						Path:     filepath.Join(bindingRoot, "binding-1A"),
						Type:     "type-1",
						Provider: "provider-1A",
						Entries: map[string]*servicebindings.Entry{
							"username": servicebindings.NewEntry(filepath.Join(bindingRoot, "binding-1A", "username")),
							"password": servicebindings.NewEntry(filepath.Join(bindingRoot, "binding-1A", "password")),
						},
					},
					servicebindings.Binding{
						Name:     "binding-1B",
						Path:     filepath.Join(bindingRoot, "binding-1B"),
						Type:     "type-1",
						Provider: "provider-1B",
						Entries: map[string]*servicebindings.Entry{
							"username": servicebindings.NewEntry(filepath.Join(bindingRoot, "binding-1B", "username")),
							"password": servicebindings.NewEntry(filepath.Join(bindingRoot, "binding-1B", "password")),
						},
					},
				))
			})

			it("resolves by type and provider (case-insensitive)", func() {
				bindings, err := resolver.Resolve("TyPe-1", "PrOvIdEr-1B", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(bindings).To(ConsistOf(
					servicebindings.Binding{
						Name:     "binding-1B",
						Path:     filepath.Join(bindingRoot, "binding-1B"),
						Type:     "type-1",
						Provider: "provider-1B",
						Entries: map[string]*servicebindings.Entry{
							"username": servicebindings.NewEntry(filepath.Join(bindingRoot, "binding-1B", "username")),
							"password": servicebindings.NewEntry(filepath.Join(bindingRoot, "binding-1B", "password")),
						},
					},
				))
			})

			it("resolves by type/provider files that contains whitespace", func() {
				bindings, err := resolver.Resolve("type-3", "provider-3", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(bindings).To(ConsistOf(
					servicebindings.Binding{
						Name:     "binding-3",
						Path:     filepath.Join(bindingRoot, "binding-3"),
						Type:     "type-3",
						Provider: "provider-3",
						Entries: map[string]*servicebindings.Entry{
							"value": servicebindings.NewEntry(filepath.Join(bindingRoot, "binding-3", "value")),
						},
					},
				))
			})

			it("allows 'metadata' as an entry name", func() {
				err := os.MkdirAll(filepath.Join(bindingRoot, "binding-metadata"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(bindingRoot, "binding-metadata", "type"), []byte("type-metadata"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(bindingRoot, "binding-metadata", "metadata"), nil, os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				bindings, err := resolver.Resolve("type-metadata", "", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(bindings).To(ConsistOf(
					servicebindings.Binding{
						Name: "binding-metadata",
						Path: filepath.Join(bindingRoot, "binding-metadata"),
						Type: "type-metadata",
						Entries: map[string]*servicebindings.Entry{
							"metadata": servicebindings.NewEntry(filepath.Join(bindingRoot, "binding-metadata", "metadata")),
						},
					},
				))
			})

			it("returns an error if type is missing", func() {
				err := os.MkdirAll(filepath.Join(bindingRoot, "bad-binding"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				_, err = resolver.Resolve("bad-type", "", "")
				Expect(err).To(MatchError(HavePrefix("failed to load bindings from '%s': failed to read binding 'bad-binding': missing 'type'", bindingRoot)))
			})

			it("allows provider to be omitted", func() {
				err := os.MkdirAll(filepath.Join(bindingRoot, "some-binding"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(bindingRoot, "some-binding", "type"), []byte("some-type"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				bindings, err := resolver.Resolve("some-type", "", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(bindings).To(ConsistOf(
					servicebindings.Binding{
						Name:     "some-binding",
						Path:     filepath.Join(bindingRoot, "some-binding"),
						Type:     "some-type",
						Provider: "",
						Entries:  map[string]*servicebindings.Entry{},
					},
				))
			})

			it("returns errors encountered reading files", func() {
				err := os.MkdirAll(filepath.Join(bindingRoot, "bad-binding"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(bindingRoot, "bad-binding", "type"), []byte("bad-type"), 000)
				Expect(err).NotTo(HaveOccurred())

				_, err = resolver.Resolve("bad-type", "", "")
				Expect(err).To(MatchError(HavePrefix("failed to load bindings from '%s': failed to read binding 'bad-binding': open %s: permission denied", bindingRoot, filepath.Join(bindingRoot, "bad-binding", "type"))))
			})

			it("returns empty list if binding root doesn't exist", func() {
				Expect(os.RemoveAll(bindingRoot)).To(Succeed())

				bindings, err := resolver.Resolve("type-1", "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(bindings).To(BeEmpty())
			})
		})

		context("ResolveOne", func() {
			it("resolves one binding (case-insensitive)", func() {
				binding, err := resolver.ResolveOne("TyPe-2", "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(binding).To(Equal(servicebindings.Binding{
					Name:     "binding-2",
					Path:     filepath.Join(bindingRoot, "binding-2"),
					Type:     "type-2",
					Provider: "provider-2",
					Entries: map[string]*servicebindings.Entry{
						"username": servicebindings.NewEntry(filepath.Join(bindingRoot, "binding-2", "username")),
						"password": servicebindings.NewEntry(filepath.Join(bindingRoot, "binding-2", "password")),
					},
				}))
			})

			it("returns an error if no matches", func() {
				_, err := resolver.ResolveOne("non-existent-type", "non-existent-provider", "")
				Expect(err).To(MatchError("found 0 bindings for type 'non-existent-type' and provider 'non-existent-provider' but expected exactly 1"))
			})

			it("returns an error if more than one match", func() {
				_, err := resolver.ResolveOne("TyPe-1", "", "")
				Expect(err).To(MatchError("found 2 bindings for type 'TyPe-1' and provider '' but expected exactly 1"))
			})
		})

		context("legacy bindings", func() {
			it("resolves legacy bindings", func() {
				err := os.MkdirAll(filepath.Join(bindingRoot, "binding-legacy", "metadata"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = os.MkdirAll(filepath.Join(bindingRoot, "binding-legacy", "secret"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(bindingRoot, "binding-legacy", "metadata", "kind"), []byte("type-legacy"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(bindingRoot, "binding-legacy", "metadata", "provider"), []byte("provider-legacy"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(bindingRoot, "binding-legacy", "metadata", "username"), nil, os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(bindingRoot, "binding-legacy", "secret", "password"), nil, os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				bindings, err := resolver.Resolve("type-legacy", "", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(bindings).To(ConsistOf(
					servicebindings.Binding{
						Name:     "binding-legacy",
						Path:     filepath.Join(bindingRoot, "binding-legacy"),
						Type:     "type-legacy",
						Provider: "provider-legacy",
						Entries: map[string]*servicebindings.Entry{
							"username": servicebindings.NewEntry(filepath.Join(bindingRoot, "binding-legacy", "metadata", "username")),
							"password": servicebindings.NewEntry(filepath.Join(bindingRoot, "binding-legacy", "secret", "password")),
						},
					},
				))
			})

			it("allows 'secret' directory to be omitted", func() {
				err := os.MkdirAll(filepath.Join(bindingRoot, "binding-legacy", "metadata"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(bindingRoot, "binding-legacy", "metadata", "kind"), []byte("type-legacy"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(bindingRoot, "binding-legacy", "metadata", "provider"), []byte("provider-legacy"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(bindingRoot, "binding-legacy", "metadata", "some-key"), nil, os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				bindings, err := resolver.Resolve("type-legacy", "", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(bindings).To(ConsistOf(
					servicebindings.Binding{
						Name:     "binding-legacy",
						Path:     filepath.Join(bindingRoot, "binding-legacy"),
						Type:     "type-legacy",
						Provider: "provider-legacy",
						Entries: map[string]*servicebindings.Entry{
							"some-key": servicebindings.NewEntry(filepath.Join(bindingRoot, "binding-legacy", "metadata", "some-key")),
						},
					},
				))
			})

			it("returns an error if kind is missing", func() {
				err := os.MkdirAll(filepath.Join(bindingRoot, "bad-binding", "metadata"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				_, err = resolver.Resolve("bad-type", "", "")
				Expect(err).To(MatchError(HavePrefix("failed to load bindings from '%s': failed to read binding 'bad-binding': missing 'kind'", bindingRoot)))
			})

			it("returns an error if provider is missing", func() {
				err := os.MkdirAll(filepath.Join(bindingRoot, "bad-binding", "metadata"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(bindingRoot, "bad-binding", "metadata", "kind"), []byte("bad-type"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				_, err = resolver.Resolve("bad-type", "", "")
				Expect(err).To(MatchError(HavePrefix("failed to load bindings from '%s': failed to read binding 'bad-binding': missing 'provider'", bindingRoot)))
			})
		})
	})
}
