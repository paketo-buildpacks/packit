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

	context("Get", func() {
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

		it("creates a sub-directory with the given name", func() {
			_, err := layers.Get("some-layer")
			Expect(err).NotTo(HaveOccurred())
			Expect(filepath.Join(layersDir, "some-layer")).To(BeADirectory())
		})

		context("failure cases", func() {
			context("when the layers directory cannot be written to", func() {
				it.Before(func() {
					Expect(os.Chmod(layersDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(layersDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := layers.Get("some-layer")
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
}
