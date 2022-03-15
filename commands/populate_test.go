package commands_test

import (
	"fmt"
	"github.com/paketo-buildpacks/packit/v2"
	production "github.com/paketo-buildpacks/packit/v2/commands"
	"github.com/sclevine/spec"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

func testCommandPopulator(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		workingDir string
		cnbDir     string
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.MkdirAll(filepath.Join(cnbDir, "bin"), os.ModePerm)).NotTo(HaveOccurred())
		Expect(ioutil.WriteFile(filepath.Join(cnbDir, "bin", "port-chooser"), []byte(""), 0644)).NotTo(HaveOccurred())

		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("PopulateExecDCommands", func() {
		var (
			layerName     string
			expectedLayer packit.Layer
		)

		it.Before(func() {
			layerName = "layer-name"

			expectedLayer = packit.Layer{
				Path:             filepath.Join(layersDir, layerName),
				Name:             layerName,
				Launch:           true,
				SharedEnv:        packit.Environment{},
				BuildEnv:         packit.Environment{},
				LaunchEnv:        packit.Environment{},
				ProcessLaunchEnv: map[string]packit.Environment{},
			}
		})

		it("does not create layerDir if no commands are provided", func() {
			layer, err := production.PopulateExecDCommands(packit.BuildContext{
				CNBPath: cnbDir,
				Layers: packit.Layers{
					Path: layersDir,
				},
			}, layerName)
			Expect(err).NotTo(HaveOccurred())

			Expect(layer).To(Equal(expectedLayer))

			_, err = os.ReadDir(filepath.Join(layersDir, layerName))
			Expect(err).To(MatchError(fs.ErrNotExist))
		})

		it("creates exec.d directory in the layer when commands are provided", func() {
			cmd1, err := ioutil.TempFile(filepath.Join(cnbDir, "bin"), "cmd1")
			Expect(err).NotTo(HaveOccurred())

			layer, err := production.PopulateExecDCommands(packit.BuildContext{
				CNBPath: cnbDir,
				Layers: packit.Layers{
					Path: layersDir,
				},
			}, layerName, filepath.Base(cmd1.Name()))
			Expect(err).NotTo(HaveOccurred())

			Expect(layer).To(Equal(expectedLayer))

			_, err = os.ReadFile(filepath.Join(layersDir, layerName, "exec.d", fmt.Sprintf("0-%s", filepath.Base(cmd1.Name()))))
			Expect(err).NotTo(HaveOccurred())
		})

		it("prepends a padded integer for lexical ordering", func() {
			N := 101

			var commands []string
			for i := 0; i < N; i++ {
				command, err := ioutil.TempFile(filepath.Join(cnbDir, "bin"), "command")
				Expect(err).NotTo(HaveOccurred())

				commands = append(commands, filepath.Base(command.Name()))
			}

			layer, err := production.PopulateExecDCommands(packit.BuildContext{
				CNBPath: cnbDir,
				Layers: packit.Layers{
					Path: layersDir,
				},
			}, layerName, commands...)
			Expect(err).NotTo(HaveOccurred())

			Expect(layer).To(Equal(expectedLayer))

			_, err = os.ReadFile(filepath.Join(layersDir, layerName, "exec.d", fmt.Sprintf("000-%s", commands[0])))
			Expect(err).NotTo(HaveOccurred())

			_, err = os.ReadFile(filepath.Join(layersDir, layerName, "exec.d", fmt.Sprintf("010-%s", commands[10])))
			Expect(err).NotTo(HaveOccurred())

			_, err = os.ReadFile(filepath.Join(layersDir, layerName, "exec.d", fmt.Sprintf("100-%s", commands[100])))
			Expect(err).NotTo(HaveOccurred())
		})

		context("failure cases", func() {
			it("throws a specific error when command executable not found", func() {
				_, err := production.PopulateExecDCommands(packit.BuildContext{
					CNBPath: cnbDir,
					Layers: packit.Layers{
						Path: layersDir,
					},
				}, layerName, "command-that-is-not-found")
				Expect(err).To(MatchError("file bin/command-that-is-not-found does not exist. Be sure to include it in the buildpack.toml"))
			})
		})
	})
}
