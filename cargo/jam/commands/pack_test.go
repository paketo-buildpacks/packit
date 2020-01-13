package commands_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/cargo/jam/commands"
	"github.com/cloudfoundry/packit/cargo/jam/commands/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPack(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer              *bytes.Buffer
		directoryDuplicator *fakes.DirectoryDuplicator
		configParser        *fakes.ConfigParser
		tarBuilder          *fakes.TarBuilder
		prePackager         *fakes.PrePackager
		fileBundler         *fakes.FileBundler
		dependencyCacher    *fakes.DependencyCacher
		tempDir             string
		tempCopyDir         string

		command commands.Pack
	)

	it.Before(func() {
		configParser = &fakes.ConfigParser{}
		configParser.ParseCall.Returns.Config = cargo.Config{
			API: "0.2",
			Buildpack: cargo.ConfigBuildpack{
				ID:   "some-buildpack-id",
				Name: "some-buildpack-name",
			},
			Metadata: cargo.ConfigMetadata{
				IncludeFiles: []string{
					"bin/build",
					"bin/detect",
					"buildpack.toml",
				},
				PrePackage: "some-prepackage-script",
				Dependencies: []cargo.ConfigMetadataDependency{
					{
						ID:      "some-dependency-id",
						Name:    "some-dependency-name",
						SHA256:  "some-dependency-sha",
						Stacks:  []string{"some-stack"},
						URI:     "some-dependency-uri",
						Version: "some-dependency-version",
					},
					{
						ID:      "other-dependency-id",
						Name:    "other-dependency-name",
						SHA256:  "other-dependency-sha",
						Stacks:  []string{"other-stack"},
						URI:     "other-dependency-uri",
						Version: "other-dependency-version",
					},
				},
			},
		}

		fileBundler = &fakes.FileBundler{}
		fileBundler.BundleCall.Returns.FileSlice = []cargo.File{
			{
				Name:       "buildpack.toml",
				Info:       cargo.NewFileInfo("buildpack.toml", len("buildpack-toml-contents"), 0644, time.Now()),
				ReadCloser: ioutil.NopCloser(strings.NewReader("buildpack-toml-contents")),
			},
		}

		directoryDuplicator = &fakes.DirectoryDuplicator{}
		tarBuilder = &fakes.TarBuilder{}
		prePackager = &fakes.PrePackager{}
		dependencyCacher = &fakes.DependencyCacher{}
		dependencyCacher.CacheCall.Returns.ConfigMetadataDependencySlice = []cargo.ConfigMetadataDependency{
			{
				ID:      "some-dependency-id",
				Name:    "some-dependency-name",
				SHA256:  "some-dependency-sha",
				Stacks:  []string{"some-stack"},
				URI:     "file:///dependencies/some-dependency-sha",
				Version: "some-dependency-version",
			},
		}

		buffer = bytes.NewBuffer(nil)

		command = commands.NewPack(directoryDuplicator, configParser, prePackager, dependencyCacher, fileBundler, tarBuilder, buffer)

		var err error
		tempDir, err = ioutil.TempDir("", "buildpack")
		Expect(err).NotTo(HaveOccurred())
		tempCopyDir, err = ioutil.TempDir("", "dup-dest")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(tempDir)).To(Succeed())
		Expect(os.RemoveAll(tempCopyDir)).To(Succeed())

	})

	context("Execute", func() {
		it("builds a buildpack", func() {
			err := command.Execute([]string{
				"--buildpack", "buildpack-root/some-buildpack.toml",
				"--version", "some-buildpack-version",
				"--output", "some-output.tgz",
				"--stack", "some-stack",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(ContainSubstring("Packing some-buildpack-name some-buildpack-version...\n"))

			Expect(directoryDuplicator.DuplicateCall.Receives.SourcePath).To(Equal("buildpack-root"))
			Expect(directoryDuplicator.DuplicateCall.Receives.DestPath).To(HavePrefix(os.TempDir()))

			buildpackRoot := directoryDuplicator.DuplicateCall.Receives.DestPath

			Expect(configParser.ParseCall.Receives.Path).To(Equal(filepath.Join(buildpackRoot, "some-buildpack.toml")))

			Expect(prePackager.ExecuteCall.Receives.Path).To(Equal("some-prepackage-script"))
			Expect(prePackager.ExecuteCall.Receives.RootDir).To(Equal(buildpackRoot))

			Expect(fileBundler.BundleCall.Receives.Path).To(Equal(buildpackRoot))
			Expect(fileBundler.BundleCall.Receives.Files).To(Equal([]string{
				"bin/build",
				"bin/detect",
				"buildpack.toml",
			}))
			Expect(fileBundler.BundleCall.Receives.Config).To(Equal(cargo.Config{
				API: "0.2",
				Buildpack: cargo.ConfigBuildpack{
					ID:      "some-buildpack-id",
					Name:    "some-buildpack-name",
					Version: "some-buildpack-version",
				},
				Metadata: cargo.ConfigMetadata{
					IncludeFiles: []string{
						"bin/build",
						"bin/detect",
						"buildpack.toml",
					},
					PrePackage: "some-prepackage-script",
					Dependencies: []cargo.ConfigMetadataDependency{
						{
							ID:      "some-dependency-id",
							Name:    "some-dependency-name",
							SHA256:  "some-dependency-sha",
							Stacks:  []string{"some-stack"},
							URI:     "some-dependency-uri",
							Version: "some-dependency-version",
						},
					},
				},
			}))

			Expect(tarBuilder.BuildCall.Receives.Path).To(Equal("some-output.tgz"))
			Expect(tarBuilder.BuildCall.Receives.Files).To(HaveLen(1))

			buildpackTOMLFile := tarBuilder.BuildCall.Receives.Files[0]
			Expect(buildpackTOMLFile.Name).To(Equal("buildpack.toml"))
			Expect(buildpackTOMLFile.Info.Size()).To(Equal(int64(len("buildpack-toml-contents"))))

			contents, err := ioutil.ReadAll(buildpackTOMLFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("buildpack-toml-contents"))

			Expect(dependencyCacher.CacheCall.CallCount).To(Equal(0))
		})

		context("when given the --offline flag", func() {
			it("downloads dependencies and builds a buildpack", func() {
				err := command.Execute([]string{
					"--buildpack", "buildpack-root/some-buildpack.toml",
					"--version", "some-buildpack-version",
					"--output", "some-output.tgz",
					"--offline",
					"--stack", "some-stack",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(buffer).To(ContainSubstring("Packing some-buildpack-name some-buildpack-version...\n"))

				Expect(directoryDuplicator.DuplicateCall.Receives.SourcePath).To(Equal("buildpack-root"))
				Expect(directoryDuplicator.DuplicateCall.Receives.DestPath).To(HavePrefix(os.TempDir()))

				buildpackRoot := directoryDuplicator.DuplicateCall.Receives.DestPath

				Expect(configParser.ParseCall.Receives.Path).To(Equal(filepath.Join(buildpackRoot, "some-buildpack.toml")))

				Expect(prePackager.ExecuteCall.Receives.Path).To(Equal("some-prepackage-script"))
				Expect(prePackager.ExecuteCall.Receives.RootDir).To(Equal(buildpackRoot))

				Expect(dependencyCacher.CacheCall.Receives.Root).To(Equal(buildpackRoot))
				Expect(dependencyCacher.CacheCall.Receives.Dependencies).To(Equal([]cargo.ConfigMetadataDependency{
					{
						ID:      "some-dependency-id",
						Name:    "some-dependency-name",
						SHA256:  "some-dependency-sha",
						Stacks:  []string{"some-stack"},
						URI:     "some-dependency-uri",
						Version: "some-dependency-version",
					},
				}))

				Expect(fileBundler.BundleCall.Receives.Path).To(Equal(buildpackRoot))
				Expect(fileBundler.BundleCall.Receives.Files).To(Equal([]string{
					"bin/build",
					"bin/detect",
					"buildpack.toml",
					"dependencies/some-dependency-sha",
				}))
				Expect(fileBundler.BundleCall.Receives.Config).To(Equal(cargo.Config{
					API: "0.2",
					Buildpack: cargo.ConfigBuildpack{
						ID:      "some-buildpack-id",
						Name:    "some-buildpack-name",
						Version: "some-buildpack-version",
					},
					Metadata: cargo.ConfigMetadata{
						IncludeFiles: []string{
							"bin/build",
							"bin/detect",
							"buildpack.toml",
							"dependencies/some-dependency-sha",
						},
						PrePackage: "some-prepackage-script",
						Dependencies: []cargo.ConfigMetadataDependency{
							{
								ID:      "some-dependency-id",
								Name:    "some-dependency-name",
								SHA256:  "some-dependency-sha",
								Stacks:  []string{"some-stack"},
								URI:     "file:///dependencies/some-dependency-sha",
								Version: "some-dependency-version",
							},
						},
					},
				}))

				Expect(tarBuilder.BuildCall.Receives.Path).To(Equal("some-output.tgz"))
				Expect(tarBuilder.BuildCall.Receives.Files).To(HaveLen(1))

				buildpackTOMLFile := tarBuilder.BuildCall.Receives.Files[0]
				Expect(buildpackTOMLFile.Name).To(Equal("buildpack.toml"))
				Expect(buildpackTOMLFile.Info.Size()).To(Equal(int64(len("buildpack-toml-contents"))))

				contents, err := ioutil.ReadAll(buildpackTOMLFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("buildpack-toml-contents"))
			})
		})
	})

	context("failure cases", func() {
		context("when given an unknown flag", func() {
			it("prints an error message", func() {
				err := command.Execute([]string{"--unknown"})
				Expect(err).To(MatchError(ContainSubstring("flag provided but not defined: -unknown")))
			})
		})
		context("when the --buildpack flag is empty", func() {
			it("prints an error message", func() {
				err := command.Execute([]string{
					"--version", "some-buildpack-version",
					"--output", filepath.Join(tempDir, "tempBuildpack.tar.gz"),
				})
				Expect(err).To(MatchError("missing required flag --buildpack"))
			})
		})

		context("when the --output flag is empty", func() {
			it("prints an error message", func() {
				err := command.Execute([]string{
					"--buildpack", filepath.Join("..", "testdata", "example-cnb", "buildpack.toml"),
					"--version", "some-buildpack-version",
				})
				Expect(err).To(MatchError("missing required flag --output"))
			})
		})

		context("when the --version flag is empty", func() {
			it("prints an error message", func() {
				err := command.Execute([]string{
					"--buildpack", filepath.Join("..", "testdata", "example-cnb", "buildpack.toml"),
					"--output", filepath.Join(tempDir, "tempBuildpack.tar.gz"),
				})
				Expect(err).To(MatchError("missing required flag --version"))
			})
		})

		context("when the buildpack parser fails", func() {
			it.Before(func() {
				configParser.ParseCall.Returns.Error = errors.New("failed to parse")
			})

			it("returns an error", func() {
				err := command.Execute([]string{
					"--buildpack", "no-such-buildpack.toml",
					"--output", filepath.Join(tempDir, "tempBuildpack.tar.gz"),
					"--version", "some-buildpack-version",
				})
				Expect(err).To(MatchError(ContainSubstring("failed to parse buildpack.toml:")))
				Expect(err).To(MatchError(ContainSubstring("failed to parse")))
			})
		})

		context("when the directoryDuplicator fails", func() {
			it.Before(func() {
				directoryDuplicator.DuplicateCall.Returns.Error = errors.New("duplication failed")
			})

			it("returns an error", func() {
				err := command.Execute([]string{
					"--buildpack", "no-such-buildpack.toml",
					"--output", filepath.Join(tempDir, "tempBuildpack.tar.gz"),
					"--version", "some-buildpack-version",
				})
				Expect(err).To(MatchError("failed to duplicate directory: duplication failed"))
			})
		})
		context("when the prepackager fails", func() {
			it.Before(func() {
				prePackager.ExecuteCall.Returns.Error = errors.New("script failed")
			})

			it("returns an error", func() {
				err := command.Execute([]string{
					"--buildpack", "no-such-buildpack.toml",
					"--output", filepath.Join(tempDir, "tempBuildpack.tar.gz"),
					"--version", "some-buildpack-version",
				})
				Expect(err).To(MatchError("failed to execute pre-packaging script \"some-prepackage-script\": script failed"))
			})
		})

		context("when the files cannot be bundled", func() {
			it.Before(func() {
				fileBundler.BundleCall.Returns.Error = errors.New("read failed")
			})

			it("returns an error", func() {
				err := command.Execute([]string{
					"--buildpack", "no-such-buildpack.toml",
					"--output", "output.tar.gz",
					"--version", "some-buildpack-version",
				})
				Expect(err).To(MatchError(ContainSubstring("failed to bundle files:")))
				Expect(err).To(MatchError(ContainSubstring("read failed")))
			})
		})

		context("when the tar builder failes", func() {
			it.Before(func() {
				tarBuilder.BuildCall.Returns.Error = errors.New("failed to build tarball")
			})

			it("returns an error", func() {
				err := command.Execute([]string{
					"--buildpack", "some-buildpack.toml",
					"--output", "some-output.tgz",
					"--version", "some-buildpack-version",
				})
				Expect(err).To(MatchError(ContainSubstring("failed to create output:")))
				Expect(err).To(MatchError(ContainSubstring("failed to build tarball")))
			})
		})
		context("when caching dependencies fails", func() {
			it("returns an error", func() {
				dependencyCacher.CacheCall.Returns.Error = errors.New("it didn't work")
				err := command.Execute([]string{
					"--buildpack", "buildpack-root/some-buildpack.toml",
					"--version", "some-buildpack-version",
					"--output", "some-output.tgz",
					"--offline",
				})
				Expect(err).To(MatchError("failed to cache dependencies: it didn't work"))
			})
		})
	})
}
