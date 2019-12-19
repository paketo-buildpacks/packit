package packit_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLayers(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir string
		layers    packit.Layers
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		layers = packit.Layers{
			Path: layersDir,
		}
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
	})

	context("layers.Get", func() {
		it("returns a layer with the given name", func() {
			layer, err := layers.Get("some-layer")
			Expect(err).NotTo(HaveOccurred())
			Expect(layer).To(Equal(packit.Layer{
				Name:      "some-layer",
				Path:      filepath.Join(layersDir, "some-layer"),
				SharedEnv: packit.NewEnvironment(),
				BuildEnv:  packit.NewEnvironment(),
				LaunchEnv: packit.NewEnvironment(),
			}))
		})

		context("when given flags", func() {
			it("applies those flags to the layer", func() {
				layer, err := layers.Get("some-layer", packit.LaunchLayer, packit.BuildLayer, packit.CacheLayer)
				Expect(err).NotTo(HaveOccurred())
				Expect(layer).To(Equal(packit.Layer{
					Name:      "some-layer",
					Path:      filepath.Join(layersDir, "some-layer"),
					Launch:    true,
					Build:     true,
					Cache:     true,
					SharedEnv: packit.NewEnvironment(),
					BuildEnv:  packit.NewEnvironment(),
					LaunchEnv: packit.NewEnvironment(),
				}))
			})
		})

		context("when the layer already exists on disk", func() {
			it.Before(func() {
				err := ioutil.WriteFile(filepath.Join(layersDir, "some-layer.toml"), []byte(`
build = true
launch = true
cache = true

[metadata]
some-key = "some-value"`), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns a layer with the existing metadata", func() {
				layer, err := layers.Get("some-layer", packit.LaunchLayer, packit.BuildLayer, packit.CacheLayer)
				Expect(err).NotTo(HaveOccurred())
				Expect(layer).To(Equal(packit.Layer{
					Name:      "some-layer",
					Path:      filepath.Join(layersDir, "some-layer"),
					Launch:    true,
					Build:     true,
					Cache:     true,
					SharedEnv: packit.NewEnvironment(),
					BuildEnv:  packit.NewEnvironment(),
					LaunchEnv: packit.NewEnvironment(),
					Metadata: map[string]interface{}{
						"some-key": "some-value",
					},
				}))
			})

			context("when the layer includes environment variable", func() {
				it.Before(func() {
					sharedEnvDir := filepath.Join(layersDir, "some-layer", "env")
					Expect(os.MkdirAll(sharedEnvDir, os.ModePerm)).To(Succeed())

					err := ioutil.WriteFile(filepath.Join(sharedEnvDir, "OVERRIDE_VAR.override"), []byte("override-value"), 0644)
					Expect(err).NotTo(HaveOccurred())

					buildEnvDir := filepath.Join(layersDir, "some-layer", "env.build")
					Expect(os.MkdirAll(buildEnvDir, os.ModePerm)).To(Succeed())

					err = ioutil.WriteFile(filepath.Join(buildEnvDir, "DEFAULT_VAR.default"), []byte("default-value"), 0644)
					Expect(err).NotTo(HaveOccurred())

					err = ioutil.WriteFile(filepath.Join(buildEnvDir, "INVALID_VAR.invalid"), []byte("invalid-value"), 0644)
					Expect(err).NotTo(HaveOccurred())

					launchEnvDir := filepath.Join(layersDir, "some-layer", "env.launch")
					Expect(os.MkdirAll(launchEnvDir, os.ModePerm)).To(Succeed())

					err = ioutil.WriteFile(filepath.Join(launchEnvDir, "APPEND_VAR.append"), []byte("append-value"), 0644)
					Expect(err).NotTo(HaveOccurred())

					err = ioutil.WriteFile(filepath.Join(launchEnvDir, "APPEND_VAR.delim"), []byte("!"), 0644)
					Expect(err).NotTo(HaveOccurred())

					err = ioutil.WriteFile(filepath.Join(launchEnvDir, "PREPEND_VAR.prepend"), []byte("prepend-value"), 0644)
					Expect(err).NotTo(HaveOccurred())

					err = ioutil.WriteFile(filepath.Join(launchEnvDir, "PREPEND_VAR.delim"), []byte("#"), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				it("returns a layer with the existing metadata", func() {
					layer, err := layers.Get("some-layer", packit.LaunchLayer, packit.BuildLayer, packit.CacheLayer)
					Expect(err).NotTo(HaveOccurred())
					Expect(layer).To(Equal(packit.Layer{
						Name:   "some-layer",
						Path:   filepath.Join(layersDir, "some-layer"),
						Launch: true,
						Build:  true,
						Cache:  true,
						SharedEnv: packit.Environment{
							"OVERRIDE_VAR.override": "override-value",
						},
						BuildEnv: packit.Environment{
							"DEFAULT_VAR.default": "default-value",
						},
						LaunchEnv: packit.Environment{
							"APPEND_VAR.append":   "append-value",
							"APPEND_VAR.delim":    "!",
							"PREPEND_VAR.prepend": "prepend-value",
							"PREPEND_VAR.delim":   "#",
						},
						Metadata: map[string]interface{}{
							"some-key": "some-value",
						},
					}))
				})
			})
		})

		context("failure cases", func() {
			context("when the layers directory contains a malformed layer toml", func() {
				it.Before(func() {
					err := ioutil.WriteFile(filepath.Join(layersDir, "some-layer.toml"), []byte("%%%"), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				it("returns an error", func() {
					_, err := layers.Get("some-layer")
					Expect(err).To(MatchError(ContainSubstring("failed to parse layer content metadata:")))
					Expect(err).To(MatchError(ContainSubstring("bare keys cannot contain '%'")))
				})
			})

			context("when the shared env directory cannot be read", func() {
				it.Before(func() {
					sharedEnvDir := filepath.Join(layersDir, "some-layer", "env")
					Expect(os.MkdirAll(sharedEnvDir, os.ModePerm)).To(Succeed())
					Expect(ioutil.WriteFile(filepath.Join(sharedEnvDir, "DEFAULT_VAR.default"), []byte("default-value"), 0000)).To(Succeed())
				})

				it("returns a layer with the existing metadata", func() {
					_, err := layers.Get("some-layer", packit.LaunchLayer, packit.BuildLayer, packit.CacheLayer)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when the build env directory cannot be read", func() {
				it.Before(func() {
					buildEnvDir := filepath.Join(layersDir, "some-layer", "env.build")
					Expect(os.MkdirAll(buildEnvDir, os.ModePerm)).To(Succeed())
					Expect(ioutil.WriteFile(filepath.Join(buildEnvDir, "DEFAULT_VAR.default"), []byte("default-value"), 0000)).To(Succeed())
				})

				it("returns a layer with the existing metadata", func() {
					_, err := layers.Get("some-layer", packit.LaunchLayer, packit.BuildLayer, packit.CacheLayer)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when the launch env directory cannot be read", func() {
				it.Before(func() {
					launchEnvDir := filepath.Join(layersDir, "some-layer", "env.launch")
					Expect(os.MkdirAll(launchEnvDir, os.ModePerm)).To(Succeed())
					Expect(ioutil.WriteFile(filepath.Join(launchEnvDir, "DEFAULT_VAR.default"), []byte("default-value"), 0000)).To(Succeed())
				})

				it("returns a layer with the existing metadata", func() {
					_, err := layers.Get("some-layer", packit.LaunchLayer, packit.BuildLayer, packit.CacheLayer)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
}
