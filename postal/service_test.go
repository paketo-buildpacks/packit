package postal_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/postal/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testService(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string

		transport       *fakes.Transport
		mappingResolver *fakes.MappingResolver

		service postal.Service
	)

	it.Before(func() {
		file, err := os.CreateTemp("", "buildpack.toml")
		Expect(err).NotTo(HaveOccurred())

		path = file.Name()
		_, err = file.WriteString(`
[[metadata.dependencies]]
deprecation_date = 2022-04-01T00:00:00Z
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack"]
uri = "some-uri"
version = "1.2.3"

[[metadata.dependencies]]
id = "some-other-entry"
sha256 = "some-other-sha"
stacks = ["some-stack"]
uri = "some-uri"
version = "1.2.4"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["other-stack"]
uri = "some-uri"
version = "1.2.5"

[[metadata.dependencies]]
id = "some-random-entry"
sha256 = "some-random-sha"
stacks = ["other-random-stack"]
uri = "some-uri"
version = "1.3.0"

[[metadata.dependencies]]
id = "some-random-other-entry"
sha256 = "some-random-other-sha"
stacks = ["some-other-random-stack"]
uri = "some-uri"
version = "2.0.0"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack"]
uri = "some-uri"
version = "4.5.6"
`)
		Expect(err).NotTo(HaveOccurred())

		Expect(file.Close()).To(Succeed())

		transport = &fakes.Transport{}

		mappingResolver = &fakes.MappingResolver{}

		service = postal.NewService(transport)

		service = service.WithDependencyMappingResolver(mappingResolver)
	})

	context("Resolve", func() {
		it("finds the best matching dependency given a plan entry", func() {
			deprecationDate, err := time.Parse(time.RFC3339, "2022-04-01T00:00:00Z")
			Expect(err).NotTo(HaveOccurred())

			dependency, err := service.Resolve(path, "some-entry", "1.2.*", "some-stack")
			Expect(err).NotTo(HaveOccurred())
			Expect(dependency).To(Equal(postal.Dependency{
				DeprecationDate: deprecationDate,
				ID:              "some-entry",
				Stacks:          []string{"some-stack"},
				URI:             "some-uri",
				SHA256:          "some-sha",
				Version:         "1.2.3",
			}))
		})

		context("when there is NOT a default version", func() {
			context("when the entry version is empty", func() {
				it("picks the dependency with the highest semantic version number", func() {
					dependency, err := service.Resolve(path, "some-entry", "", "some-stack")
					Expect(err).NotTo(HaveOccurred())
					Expect(dependency).To(Equal(postal.Dependency{
						ID:      "some-entry",
						Stacks:  []string{"some-stack"},
						URI:     "some-uri",
						SHA256:  "some-sha",
						Version: "4.5.6",
					}))
				})
			})

			context("when the entry version is default", func() {
				it("picks the dependency with the highest semantic version number", func() {
					dependency, err := service.Resolve(path, "some-entry", "default", "some-stack")
					Expect(err).NotTo(HaveOccurred())
					Expect(dependency).To(Equal(postal.Dependency{
						ID:      "some-entry",
						Stacks:  []string{"some-stack"},
						URI:     "some-uri",
						SHA256:  "some-sha",
						Version: "4.5.6",
					}))
				})
			})

			context("when there is a version with a major, minor, patch, and pessimistic operator (~>)", func() {
				it("picks the dependency >= version and < major.minor+1", func() {
					deprecationDate, err := time.Parse(time.RFC3339, "2022-04-01T00:00:00Z")
					Expect(err).NotTo(HaveOccurred())

					dependency, err := service.Resolve(path, "some-entry", "~> 1.2.0", "some-stack")
					Expect(err).NotTo(HaveOccurred())
					Expect(dependency).To(Equal(postal.Dependency{
						DeprecationDate: deprecationDate,
						ID:              "some-entry",
						Stacks:          []string{"some-stack"},
						URI:             "some-uri",
						SHA256:          "some-sha",
						Version:         "1.2.3",
					}))
				})
			})

			context("when there is a version with a major, minor, and pessimistic operator (~>)", func() {
				it("picks the dependency >= version and < major+1", func() {
					deprecationDate, err := time.Parse(time.RFC3339, "2022-04-01T00:00:00Z")
					Expect(err).NotTo(HaveOccurred())

					dependency, err := service.Resolve(path, "some-entry", "~> 1.1", "some-stack")
					Expect(err).NotTo(HaveOccurred())
					Expect(dependency).To(Equal(postal.Dependency{
						DeprecationDate: deprecationDate,
						ID:              "some-entry",
						Stacks:          []string{"some-stack"},
						URI:             "some-uri",
						SHA256:          "some-sha",
						Version:         "1.2.3",
					}))
				})
			})

			context("when there is a version with a major line only and pessimistic operator (~>)", func() {
				it("picks the dependency >= version.0.0 and < major+1.0.0", func() {
					deprecationDate, err := time.Parse(time.RFC3339, "2022-04-01T00:00:00Z")
					Expect(err).NotTo(HaveOccurred())

					dependency, err := service.Resolve(path, "some-entry", "~> 1", "some-stack")
					Expect(err).NotTo(HaveOccurred())
					Expect(dependency).To(Equal(postal.Dependency{
						DeprecationDate: deprecationDate,
						ID:              "some-entry",
						Stacks:          []string{"some-stack"},
						URI:             "some-uri",
						SHA256:          "some-sha",
						Version:         "1.2.3",
					}))
				})
			})
		})

		context("when there is a default version", func() {
			it.Before(func() {
				err := os.WriteFile(path, []byte(`
[metadata]
[metadata.default-versions]
some-entry = "1.2.x"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack"]
uri = "some-uri"
version = "1.2.3"

[[metadata.dependencies]]
id = "some-other-entry"
sha256 = "some-other-sha"
stacks = ["some-stack"]
uri = "some-uri"
version = "1.2.4"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["other-stack"]
uri = "some-uri"
version = "1.2.5"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack"]
uri = "some-uri"
version = "4.5.6"
`), 0600)
				Expect(err).NotTo(HaveOccurred())
			})

			context("when the entry version is empty", func() {
				it("picks the dependency that best matches the default version", func() {
					dependency, err := service.Resolve(path, "some-entry", "", "some-stack")
					Expect(err).NotTo(HaveOccurred())
					Expect(dependency).To(Equal(postal.Dependency{
						ID:      "some-entry",
						Stacks:  []string{"some-stack"},
						URI:     "some-uri",
						SHA256:  "some-sha",
						Version: "1.2.3",
					}))
				})
			})

			context("when the entry version is default", func() {
				it("picks the dependency that best matches the default version", func() {
					dependency, err := service.Resolve(path, "some-entry", "default", "some-stack")
					Expect(err).NotTo(HaveOccurred())
					Expect(dependency).To(Equal(postal.Dependency{
						ID:      "some-entry",
						Stacks:  []string{"some-stack"},
						URI:     "some-uri",
						SHA256:  "some-sha",
						Version: "1.2.3",
					}))
				})
			})
		})

		context("failure cases", func() {
			context("when the buildpack.toml is malformed", func() {
				it.Before(func() {
					err := os.WriteFile(path, []byte("this is not toml"), 0600)
					Expect(err).NotTo(HaveOccurred())
				})

				it("returns an error", func() {
					_, err := service.Resolve(path, "some-entry", "1.2.3", "some-stack")
					Expect(err).To(MatchError(ContainSubstring("failed to parse buildpack.toml")))
				})
			})

			context("when the entry version constraint is not valid", func() {
				it("returns an error", func() {
					_, err := service.Resolve(path, "some-entry", "this-is-not-semver", "some-stack")
					Expect(err).To(MatchError(ContainSubstring("improper constraint")))
				})
			})

			context("when the dependency version is not valid", func() {
				it.Before(func() {
					err := os.WriteFile(path, []byte(`
[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack"]
uri = "some-uri"
version = "this is super not semver"
`), 0600)
					Expect(err).NotTo(HaveOccurred())
				})

				it("returns an error", func() {
					_, err := service.Resolve(path, "some-entry", "1.2.3", "some-stack")
					Expect(err).To(MatchError(ContainSubstring("Invalid Semantic Version")))
				})
			})

			context("when the entry version constraint cannot be satisfied", func() {
				it("returns an error with all the supported versions listed", func() {
					_, err := service.Resolve(path, "some-entry", "9.9.9", "some-stack")
					Expect(err).To(MatchError(ContainSubstring("failed to satisfy \"some-entry\" dependency version constraint \"9.9.9\": no compatible versions. Supported versions are: [1.2.3, 4.5.6]")))
				})
			})
		})
	})

	context("Deliver", func() {
		var (
			dependencySHA string
			layerPath     string
			platformPath  string
			deliver       func() error
		)

		it.Before(func() {
			var err error
			layerPath, err = os.MkdirTemp("", "layer")
			Expect(err).NotTo(HaveOccurred())

			platformPath, err = os.MkdirTemp("", "platform")
			Expect(err).NotTo(HaveOccurred())

			buffer := bytes.NewBuffer(nil)
			zw := gzip.NewWriter(buffer)
			tw := tar.NewWriter(zw)

			Expect(tw.WriteHeader(&tar.Header{Name: "./some-dir", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
			_, err = tw.Write(nil)
			Expect(err).NotTo(HaveOccurred())

			nestedFile := "./some-dir/some-file"
			Expect(tw.WriteHeader(&tar.Header{Name: nestedFile, Mode: 0755, Size: int64(len(nestedFile))})).To(Succeed())
			_, err = tw.Write([]byte(nestedFile))
			Expect(err).NotTo(HaveOccurred())

			for _, file := range []string{"./first", "./second", "./third"} {
				Expect(tw.WriteHeader(&tar.Header{Name: file, Mode: 0755, Size: int64(len(file))})).To(Succeed())
				_, err = tw.Write([]byte(file))
				Expect(err).NotTo(HaveOccurred())
			}

			linkName := "./symlink"
			linkDest := "./first"
			Expect(tw.WriteHeader(&tar.Header{Name: linkName, Mode: 0777, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: linkDest})).To(Succeed())
			// what does a sylink actually look like??
			_, err = tw.Write([]byte{})
			Expect(err).NotTo(HaveOccurred())
			// add a symlink header

			Expect(tw.Close()).To(Succeed())
			Expect(zw.Close()).To(Succeed())

			sum := sha256.Sum256(buffer.Bytes())
			dependencySHA = hex.EncodeToString(sum[:])

			transport.DropCall.Returns.ReadCloser = io.NopCloser(buffer)

			deliver = func() error {
				return service.Deliver(postal.Dependency{
					ID:      "some-entry",
					Stacks:  []string{"some-stack"},
					URI:     "some-entry.tgz",
					SHA256:  dependencySHA,
					Version: "1.2.3",
				}, "some-cnb-path",
					layerPath,
					platformPath,
				)
			}
		})

		it.After(func() {
			Expect(os.RemoveAll(layerPath)).To(Succeed())
		})

		it("downloads the dependency and unpackages it into the path", func() {
			err := deliver()

			Expect(err).NotTo(HaveOccurred())

			Expect(transport.DropCall.Receives.Root).To(Equal("some-cnb-path"))
			Expect(transport.DropCall.Receives.Uri).To(Equal("some-entry.tgz"))

			files, err := filepath.Glob(fmt.Sprintf("%s/*", layerPath))
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(ConsistOf([]string{
				filepath.Join(layerPath, "first"),
				filepath.Join(layerPath, "second"),
				filepath.Join(layerPath, "third"),
				filepath.Join(layerPath, "some-dir"),
				filepath.Join(layerPath, "symlink"),
			}))

			info, err := os.Stat(filepath.Join(layerPath, "first"))
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode()).To(Equal(os.FileMode(0755)))
		})

		context("when there is a dependency mapping via binding", func() {
			it.Before(func() {
				mappingResolver.FindDependencyMappingCall.Returns.String = "dependency-mapping-entry.tgz"
			})

			it("looks up the dependency from the platform binding and downloads that instead", func() {
				err := deliver()

				Expect(err).NotTo(HaveOccurred())

				Expect(mappingResolver.FindDependencyMappingCall.Receives.SHA256).To(Equal(dependencySHA))
				Expect(mappingResolver.FindDependencyMappingCall.Receives.BindingPath).To(Equal(filepath.Join(platformPath, "bindings")))
				Expect(transport.DropCall.Receives.Root).To(Equal("some-cnb-path"))
				Expect(transport.DropCall.Receives.Uri).To(Equal("dependency-mapping-entry.tgz"))

				files, err := filepath.Glob(fmt.Sprintf("%s/*", layerPath))
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(ConsistOf([]string{
					filepath.Join(layerPath, "first"),
					filepath.Join(layerPath, "second"),
					filepath.Join(layerPath, "third"),
					filepath.Join(layerPath, "some-dir"),
					filepath.Join(layerPath, "symlink"),
				}))

				info, err := os.Stat(filepath.Join(layerPath, "first"))
				Expect(err).NotTo(HaveOccurred())
				Expect(info.Mode()).To(Equal(os.FileMode(0755)))
			})
		})

		context("failure cases", func() {
			context("when the transport cannot fetch a dependency", func() {
				it.Before(func() {
					transport.DropCall.Returns.Error = errors.New("there was an error")
				})

				it("returns an error", func() {
					err := deliver()

					Expect(err).To(MatchError("failed to fetch dependency: there was an error"))
				})
			})

			context("when the file contents are empty", func() {
				it.Before(func() {
					// This is a FLAC header
					buffer := bytes.NewBuffer([]byte("\x66\x4C\x61\x43\x00\x00\x00\x22"))
					transport.DropCall.Returns.ReadCloser = io.NopCloser(buffer)

					sum := sha256.Sum256(buffer.Bytes())
					dependencySHA = hex.EncodeToString(sum[:])
				})

				it("fails to create a gzip reader", func() {
					err := deliver()

					Expect(err).To(MatchError(ContainSubstring("unsupported archive type")))
				})
			})

			context("when the file contents are malformed", func() {
				it.Before(func() {
					buffer := bytes.NewBuffer(nil)
					gzipWriter := gzip.NewWriter(buffer)

					_, err := gzipWriter.Write([]byte("something"))
					Expect(err).NotTo(HaveOccurred())

					Expect(gzipWriter.Close()).To(Succeed())

					transport.DropCall.Returns.ReadCloser = io.NopCloser(buffer)

					sum := sha256.Sum256(buffer.Bytes())
					dependencySHA = hex.EncodeToString(sum[:])
				})

				it("fails to create a tar reader", func() {
					err := deliver()

					Expect(err).To(MatchError(ContainSubstring("failed to read tar response")))
				})
			})

			context("when the file checksum does not match", func() {
				it("fails to create a tar reader", func() {
					err := service.Deliver(postal.Dependency{
						ID:      "some-entry",
						Stacks:  []string{"some-stack"},
						URI:     "some-entry.tgz",
						SHA256:  "this is not a valid checksum",
						Version: "1.2.3",
					}, "some-cnb-path",
						layerPath,
						platformPath,
					)

					Expect(err).To(MatchError(ContainSubstring("checksum does not match")))
				})
			})

			context("when it does not have permission to write into directory on container", func() {
				it.Before(func() {
					Expect(os.Chmod(layerPath, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(layerPath, 0755)).To(Succeed())
				})

				it("fails to make a dir", func() {
					err := deliver()

					Expect(err).To(MatchError(ContainSubstring("failed to create archived directory")))
				})
			})

			context("when it does not have permission to write into directory that it decompressed", func() {
				var testDir string
				it.Before(func() {
					testDir = filepath.Join(layerPath, "some-dir")
					Expect(os.MkdirAll(testDir, os.ModePerm)).To(Succeed())
					Expect(os.Chmod(testDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(testDir, 0755)).To(Succeed())
				})

				it("fails to make a file", func() {
					err := deliver()

					Expect(err).To(MatchError(ContainSubstring("failed to create archived file")))
				})
			})

			context("when it is given a broken symlink", func() {
				it.Before(func() {
					buffer := bytes.NewBuffer(nil)
					zw := gzip.NewWriter(buffer)
					tw := tar.NewWriter(zw)

					linkName := "./symlink"
					Expect(tw.WriteHeader(&tar.Header{Name: linkName, Mode: 0777, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: ""})).To(Succeed())
					// what does a sylink actually look like??
					_, err := tw.Write([]byte{})
					Expect(err).NotTo(HaveOccurred())
					// add a symlink header

					Expect(tw.Close()).To(Succeed())
					Expect(zw.Close()).To(Succeed())

					sum := sha256.Sum256(buffer.Bytes())
					dependencySHA = hex.EncodeToString(sum[:])

					transport.DropCall.Returns.ReadCloser = io.NopCloser(buffer)
				})

				it("fails to extract the symlink", func() {
					err := deliver()

					Expect(err).To(MatchError(ContainSubstring("failed to extract symlink")))
				})
			})
		})
	})

	context("Install", func() {
		var (
			dependencySHA string
			layerPath     string
			install       func() error
		)

		it.Before(func() {
			var err error
			layerPath, err = os.MkdirTemp("", "layer")
			Expect(err).NotTo(HaveOccurred())

			buffer := bytes.NewBuffer(nil)
			zw := gzip.NewWriter(buffer)
			tw := tar.NewWriter(zw)

			Expect(tw.WriteHeader(&tar.Header{Name: "./some-dir", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
			_, err = tw.Write(nil)
			Expect(err).NotTo(HaveOccurred())

			nestedFile := "./some-dir/some-file"
			Expect(tw.WriteHeader(&tar.Header{Name: nestedFile, Mode: 0755, Size: int64(len(nestedFile))})).To(Succeed())
			_, err = tw.Write([]byte(nestedFile))
			Expect(err).NotTo(HaveOccurred())

			for _, file := range []string{"./first", "./second", "./third"} {
				Expect(tw.WriteHeader(&tar.Header{Name: file, Mode: 0755, Size: int64(len(file))})).To(Succeed())
				_, err = tw.Write([]byte(file))
				Expect(err).NotTo(HaveOccurred())
			}

			linkName := "./symlink"
			linkDest := "./first"
			Expect(tw.WriteHeader(&tar.Header{Name: linkName, Mode: 0777, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: linkDest})).To(Succeed())
			// what does a sylink actually look like??
			_, err = tw.Write([]byte{})
			Expect(err).NotTo(HaveOccurred())
			// add a symlink header

			Expect(tw.Close()).To(Succeed())
			Expect(zw.Close()).To(Succeed())

			sum := sha256.Sum256(buffer.Bytes())
			dependencySHA = hex.EncodeToString(sum[:])

			transport.DropCall.Returns.ReadCloser = io.NopCloser(buffer)

			install = func() error {
				return service.Install(postal.Dependency{
					ID:      "some-entry",
					Stacks:  []string{"some-stack"},
					URI:     "some-entry.tgz",
					SHA256:  dependencySHA,
					Version: "1.2.3",
				}, "some-cnb-path",
					layerPath,
				)
			}
		})

		it.After(func() {
			Expect(os.RemoveAll(layerPath)).To(Succeed())
		})

		it("downloads the dependency and unpackages it into the path", func() {
			err := install()

			Expect(err).NotTo(HaveOccurred())

			Expect(transport.DropCall.Receives.Root).To(Equal("some-cnb-path"))
			Expect(transport.DropCall.Receives.Uri).To(Equal("some-entry.tgz"))

			files, err := filepath.Glob(fmt.Sprintf("%s/*", layerPath))
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(ConsistOf([]string{
				filepath.Join(layerPath, "first"),
				filepath.Join(layerPath, "second"),
				filepath.Join(layerPath, "third"),
				filepath.Join(layerPath, "some-dir"),
				filepath.Join(layerPath, "symlink"),
			}))

			info, err := os.Stat(filepath.Join(layerPath, "first"))
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode()).To(Equal(os.FileMode(0755)))
		})

		context("when there is a dependency mapping via binding", func() {
			it.Before(func() {
				mappingResolver.FindDependencyMappingCall.Returns.String = "dependency-mapping-entry.tgz"
			})

			it("looks up the dependency from the platform binding and downloads that instead", func() {
				err := install()

				Expect(err).NotTo(HaveOccurred())

				Expect(mappingResolver.FindDependencyMappingCall.Receives.SHA256).To(Equal(dependencySHA))
				Expect(mappingResolver.FindDependencyMappingCall.Receives.BindingPath).To(Equal("/platform/bindings"))
				Expect(transport.DropCall.Receives.Root).To(Equal("some-cnb-path"))
				Expect(transport.DropCall.Receives.Uri).To(Equal("dependency-mapping-entry.tgz"))

				files, err := filepath.Glob(fmt.Sprintf("%s/*", layerPath))
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(ConsistOf([]string{
					filepath.Join(layerPath, "first"),
					filepath.Join(layerPath, "second"),
					filepath.Join(layerPath, "third"),
					filepath.Join(layerPath, "some-dir"),
					filepath.Join(layerPath, "symlink"),
				}))

				info, err := os.Stat(filepath.Join(layerPath, "first"))
				Expect(err).NotTo(HaveOccurred())
				Expect(info.Mode()).To(Equal(os.FileMode(0755)))
			})
		})

		context("failure cases", func() {
			context("when the transport cannot fetch a dependency", func() {
				it.Before(func() {
					transport.DropCall.Returns.Error = errors.New("there was an error")
				})

				it("returns an error", func() {
					err := install()

					Expect(err).To(MatchError("failed to fetch dependency: there was an error"))
				})
			})

			context("when the file contents are empty", func() {
				it.Before(func() {
					// This is a FLAC header
					buffer := bytes.NewBuffer([]byte("\x66\x4C\x61\x43\x00\x00\x00\x22"))
					transport.DropCall.Returns.ReadCloser = io.NopCloser(buffer)

					sum := sha256.Sum256(buffer.Bytes())
					dependencySHA = hex.EncodeToString(sum[:])
				})

				it("fails to create a gzip reader", func() {
					err := install()

					Expect(err).To(MatchError(ContainSubstring("unsupported archive type")))
				})
			})

			context("when the file contents are malformed", func() {
				it.Before(func() {
					buffer := bytes.NewBuffer(nil)
					gzipWriter := gzip.NewWriter(buffer)

					_, err := gzipWriter.Write([]byte("something"))
					Expect(err).NotTo(HaveOccurred())

					Expect(gzipWriter.Close()).To(Succeed())

					transport.DropCall.Returns.ReadCloser = io.NopCloser(buffer)

					sum := sha256.Sum256(buffer.Bytes())
					dependencySHA = hex.EncodeToString(sum[:])
				})

				it("fails to create a tar reader", func() {
					err := install()

					Expect(err).To(MatchError(ContainSubstring("failed to read tar response")))
				})
			})

			context("when the file checksum does not match", func() {
				it("fails to create a tar reader", func() {
					err := service.Install(postal.Dependency{
						ID:      "some-entry",
						Stacks:  []string{"some-stack"},
						URI:     "some-entry.tgz",
						SHA256:  "this is not a valid checksum",
						Version: "1.2.3",
					}, "some-cnb-path",
						layerPath,
					)

					Expect(err).To(MatchError(ContainSubstring("checksum does not match")))
				})
			})

			context("when it does not have permission to write into directory on container", func() {
				it.Before(func() {
					Expect(os.Chmod(layerPath, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(layerPath, 0755)).To(Succeed())
				})

				it("fails to make a dir", func() {
					err := install()

					Expect(err).To(MatchError(ContainSubstring("failed to create archived directory")))
				})
			})

			context("when it does not have permission to write into directory that it decompressed", func() {
				var testDir string
				it.Before(func() {
					testDir = filepath.Join(layerPath, "some-dir")
					Expect(os.MkdirAll(testDir, os.ModePerm)).To(Succeed())
					Expect(os.Chmod(testDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(testDir, 0755)).To(Succeed())
				})

				it("fails to make a file", func() {
					err := install()

					Expect(err).To(MatchError(ContainSubstring("failed to create archived file")))
				})
			})

			context("when it is given a broken symlink", func() {
				it.Before(func() {
					buffer := bytes.NewBuffer(nil)
					zw := gzip.NewWriter(buffer)
					tw := tar.NewWriter(zw)

					linkName := "./symlink"
					Expect(tw.WriteHeader(&tar.Header{Name: linkName, Mode: 0777, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: ""})).To(Succeed())
					// what does a sylink actually look like??
					_, err := tw.Write([]byte{})
					Expect(err).NotTo(HaveOccurred())
					// add a symlink header

					Expect(tw.Close()).To(Succeed())
					Expect(zw.Close()).To(Succeed())

					sum := sha256.Sum256(buffer.Bytes())
					dependencySHA = hex.EncodeToString(sum[:])

					transport.DropCall.Returns.ReadCloser = io.NopCloser(buffer)
				})

				it("fails to extract the symlink", func() {
					err := install()

					Expect(err).To(MatchError(ContainSubstring("failed to extract symlink")))
				})
			})
		})
	})

	context("GenerateBillOfMaterials", func() {
		var deprecationDate time.Time

		it.Before(func() {
			var err error
			deprecationDate, err = time.Parse(time.RFC3339, "2022-04-01T00:00:00Z")
			Expect(err).NotTo(HaveOccurred())
		})

		it("returns a list of BOMEntry values", func() {
			entries := service.GenerateBillOfMaterials(
				postal.Dependency{
					DeprecationDate: deprecationDate,
					ID:              "some-entry",
					Name:            "Some Entry",
					SHA256:          "some-sha",
					Source:          "some-source",
					Stacks:          []string{"some-stack"},
					URI:             "some-uri",
					Version:         "1.2.3",
				},
				postal.Dependency{
					ID:      "other-entry",
					Name:    "Other Entry",
					SHA256:  "other-sha",
					Source:  "other-source",
					Stacks:  []string{"other-stack"},
					URI:     "other-uri",
					Version: "4.5.6",
				},
			)
			Expect(entries).To(Equal([]packit.BOMEntry{
				{
					Name: "Some Entry",
					Metadata: map[string]interface{}{
						"deprecation-date": deprecationDate,
						"sha256":           "some-sha",
						"stacks":           []string{"some-stack"},
						"uri":              "some-uri",
						"version":          "1.2.3",
					},
				},
				{
					Name: "Other Entry",
					Metadata: map[string]interface{}{
						"sha256":  "other-sha",
						"stacks":  []string{"other-stack"},
						"uri":     "other-uri",
						"version": "4.5.6",
					},
				},
			}))
		})
	})
}
