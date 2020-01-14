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

func testLayer(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir string
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
	})

	context("Reset", func() {
		var layer packit.Layer

		context("when there is no previous build", func() {
			it.Before(func() {
				layer = packit.Layer{
					Name:   "some-layer",
					Path:   filepath.Join(layersDir, "some-layer"),
					Launch: true,
					Build:  true,
					Cache:  true,
				}
			})

			it("initializes an empty layer", func() {
				Expect(layer.Reset()).To(Succeed())

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

				Expect(filepath.Join(layersDir, "some-layer")).To(BeADirectory())
			})
		})

		context("when cache is retrieved from previous build", func() {
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

				layer = packit.Layer{
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
						"APPEND_VAR.append": "append-value",
						"APPEND_VAR.delim":  "!",
					},
					Metadata: map[string]interface{}{
						"some-key": "some-value",
					},
				}
			})

			context("when Reset is called on a layer", func() {
				it("resets all of the layer data and clears the directory", func() {
					err := layer.Reset()
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

					Expect(filepath.Join(layersDir, "some-layer")).To(BeADirectory())

					files, err := filepath.Glob(filepath.Join(layersDir, "some-layer", "*"))
					Expect(err).NotTo(HaveOccurred())

					Expect(files).To(BeEmpty())
				})
			})
		})

		context("failure cases", func() {
			context("could not remove files in layer", func() {
				it.Before(func() {
					Expect(os.Chmod(layersDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(layersDir, 0777)).To(Succeed())
				})

				it("return an error", func() {
					layer := packit.Layer{
						Name: "some-layer",
						Path: filepath.Join(layersDir, "some-layer"),
					}

					err := layer.Reset()
					Expect(err).To(MatchError(ContainSubstring("error could not remove file: ")))
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
}
