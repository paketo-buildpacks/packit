package postal_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/postal/fakes"
	"github.com/sclevine/spec"

	//nolint Ignore SA1019, usage of deprecated package within a deprecated test case
	"github.com/paketo-buildpacks/packit/v2/paketosbom"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func testService(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string

		transport       *fakes.Transport
		mappingResolver *fakes.MappingResolver
		mirrorResolver  *fakes.MirrorResolver

		service postal.Service
	)

	it.Before(func() {
		file, err := os.CreateTemp("", "buildpack.toml")
		Expect(err).NotTo(HaveOccurred())

		path = file.Name()
		_, err = file.WriteString(`
[[metadata.dependencies]]
deprecation_date = 2022-04-01T00:00:00Z
cpe = "some-cpe"
cpes = ["some-cpe", "other-cpe"]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack"]
uri = "some-uri"
version = "1.2.3"

[[metadata.dependencies]]
id = "some-other-entry"
cpes = ["some-cpe", "other-cpe"]
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
cpe = "some-cpe"
cpes = ["some-cpe", "other-cpe"]
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
strip-components = 1

[[metadata.dependencies]]
id = "some-other-entry"
sha256 = "some-sha"
stacks = ["*"]
uri = "some-uri"
version = "4.5.6"
strip-components = 1
`)
		Expect(err).NotTo(HaveOccurred())

		Expect(file.Close()).To(Succeed())

		transport = &fakes.Transport{}

		mappingResolver = &fakes.MappingResolver{}

		mirrorResolver = &fakes.MirrorResolver{}

		service = postal.NewService(transport).
			WithDependencyMappingResolver(mappingResolver).
			WithDependencyMirrorResolver(mirrorResolver)

		Expect(os.Setenv("BP_ARCH", "amd64")).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.Unsetenv("BP_ARCH")).To(Succeed())
	})

	context("Resolve", func() {
		it("finds the best matching dependency given a plan entry", func() {
			deprecationDate, err := time.Parse(time.RFC3339, "2022-04-01T00:00:00Z")
			Expect(err).NotTo(HaveOccurred())

			dependency, err := service.Resolve(path, "some-entry", "1.2.*", "some-stack")
			Expect(err).NotTo(HaveOccurred())
			Expect(dependency).To(Equal(postal.Dependency{
				CPE:             "some-cpe",
				CPEs:            []string{"some-cpe", "other-cpe"},
				DeprecationDate: deprecationDate,
				ID:              "some-entry",
				Stacks:          []string{"some-stack"},
				URI:             "some-uri",
				SHA256:          "some-sha",
				Version:         "1.2.3",
			}))
		})

		context("when the dependency has a wildcard stack", func() {
			it("is compatible with all stack ids", func() {
				dependency, err := service.Resolve(path, "some-other-entry", "", "random-stack")
				Expect(err).NotTo(HaveOccurred())
				Expect(dependency).To(Equal(postal.Dependency{
					ID:              "some-other-entry",
					Stacks:          []string{"*"},
					URI:             "some-uri",
					SHA256:          "some-sha",
					Version:         "4.5.6",
					StripComponents: 1,
				}))
			})
		})

		context("when there is NOT a default version", func() {
			context("when the entry version is empty", func() {
				it("picks the dependency with the highest semantic version number", func() {
					dependency, err := service.Resolve(path, "some-entry", "", "some-stack")
					Expect(err).NotTo(HaveOccurred())
					Expect(dependency).To(Equal(postal.Dependency{
						ID:              "some-entry",
						Stacks:          []string{"some-stack"},
						URI:             "some-uri",
						SHA256:          "some-sha",
						Version:         "4.5.6",
						StripComponents: 1,
					}))
				})
			})

			context("when the entry version is default", func() {
				it("picks the dependency with the highest semantic version number", func() {
					dependency, err := service.Resolve(path, "some-entry", "default", "some-stack")
					Expect(err).NotTo(HaveOccurred())
					Expect(dependency).To(Equal(postal.Dependency{
						ID:              "some-entry",
						Stacks:          []string{"some-stack"},
						URI:             "some-uri",
						SHA256:          "some-sha",
						Version:         "4.5.6",
						StripComponents: 1,
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
						CPE:             "some-cpe",
						CPEs:            []string{"some-cpe", "other-cpe"},
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
						CPE:             "some-cpe",
						CPEs:            []string{"some-cpe", "other-cpe"},
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
						CPE:             "some-cpe",
						CPEs:            []string{"some-cpe", "other-cpe"},
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

		context("when there architecture specific versions", func() {
			it.Before(func() {
				err := os.WriteFile(path, []byte(`
[metadata]
[metadata.default-versions]
some-entry = "1.2.x"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha-amd64"
stacks = ["*"]
uri = "some-uri"
version = "1.2.3"
arch = "amd64"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha-arm"
stacks = ["*"]
uri = "some-uri"
version = "1.2.3"
arch = "arm"

[[metadata.dependencies]]
id = "some-other-entry"
sha256 = "some-other-sha"
stacks = ["*"]
uri = "some-uri"
version = "1.2.4"

`), 0600)
				Expect(err).NotTo(HaveOccurred())
			})

			context("BP_ARCH=amd64", func() {
				it.Before(func() {
					Expect(os.Setenv("BP_ARCH", "amd64")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("BP_ARCH")).To(Succeed())
				})

				it("picks the dependency with the correct arch", func() {
					dependency, err := service.Resolve(path, "some-entry", "1.2.3", "some-stack")
					Expect(err).NotTo(HaveOccurred())
					Expect(dependency.SHA256).To(Equal("some-sha-amd64"))
				})

				it("picks the dependency with the no arch", func() {
					dependency, err := service.Resolve(path, "some-other-entry", "1.2.4", "some-stack")
					Expect(err).NotTo(HaveOccurred())
					Expect(dependency.SHA256).To(Equal("some-other-sha"))
				})
			})

			context("BP_ARCH=arm", func() {
				it.Before(func() {
					Expect(os.Setenv("BP_ARCH", "arm")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("BP_ARCH")).To(Succeed())
				})

				it("picks the dependency with the correct arch", func() {
					dependency, err := service.Resolve(path, "some-entry", "1.2.3", "some-stack")
					Expect(err).NotTo(HaveOccurred())
					Expect(dependency.SHA256).To(Equal("some-sha-arm"))
				})

				it("fails if no dependency with the correct arch is found", func() {
					_, err := service.Resolve(path, "some-other-entry", "1.2.4", "some-stack")
					Expect(err).To(MatchError(ContainSubstring("failed to satisfy")))
					Expect(err).To(MatchError(ContainSubstring("\"arm\"")))
				})
			})
		})

		context("when both a wildcard stack constraint and a specific stack constraint exist for the same dependency version", func() {
			it.Before(func() {
				err := os.WriteFile(path, []byte(`
[metadata]
[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack"]
uri = "some-uri-specific-stack"
version = "1.2.1"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["*"]
uri = "some-uri-only-wildcard"
version = "1.2.1"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack","*"]
uri = "some-uri-only-wildcard"
version = "1.2.3"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack"]
uri = "some-uri-specific-stack"
version = "1.2.3"
`), 0600)

				Expect(err).NotTo(HaveOccurred())
			})

			it("selects the more specific stack constraint", func() {
				dependency, err := service.Resolve(path, "some-entry", "*", "some-stack")
				Expect(err).NotTo(HaveOccurred())
				Expect(dependency).To(Equal(postal.Dependency{
					ID:      "some-entry",
					Stacks:  []string{"some-stack"},
					URI:     "some-uri-specific-stack",
					SHA256:  "some-sha",
					Version: "1.2.3",
				}))
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

			context("when multiple dependencies have a wildcard stack for the same version", func() {
				it.Before(func() {
					err := os.WriteFile(path, []byte(`
[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha-A"
stacks = ["some-stack","*"]
uri = "some-uri-A"
version = "1.2.3"

[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha-B"
stacks = ["some-stack","some-other-stack","*"]
uri = "some-uri-B"
version = "1.2.3"
`), 0600)
					Expect(err).NotTo(HaveOccurred())
				})

				it("returns an error", func() {
					_, err := service.Resolve(path, "some-entry", "1.2.3", "some-stack")
					Expect(err).To(MatchError(ContainSubstring(`multiple dependencies support wildcard stack for version: "1.2.3"`)))
				})
			})

			context("when the entry version constraint cannot be satisfied", func() {
				it("returns a typed error with all the supported versions listed", func() {
					expectedErr := &postal.ErrNoDeps{}
					_, err := service.Resolve(path, "some-entry", "9.9.9", "some-stack")
					Expect(errors.As(err, &expectedErr)).To(BeTrue())
					Expect(err).To(MatchError(ContainSubstring("failed to satisfy \"some-entry\" dependency version constraint \"9.9.9\": no compatible versions on \"some-stack\" stack with architecture \"amd64\". Supported versions are: [1.2.3, 4.5.6]")))
				})
			})
		})
	})

	context("Deliver", func() {
		var (
			dependencyHash string
			hash512        string
			layerPath      string
			deliver        func() error
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
			_, err = tw.Write([]byte{})
			Expect(err).NotTo(HaveOccurred())

			Expect(tw.Close()).To(Succeed())
			Expect(zw.Close()).To(Succeed())

			sum := sha256.Sum256(buffer.Bytes())
			dependencyHash = hex.EncodeToString(sum[:])

			sum512 := sha512.Sum512(buffer.Bytes())
			hash512 = hex.EncodeToString(sum512[:])

			transport.DropCall.Returns.ReadCloser = io.NopCloser(buffer)

			deliver = func() error {
				return service.Deliver(
					postal.Dependency{
						ID:      "some-entry",
						Stacks:  []string{"some-stack"},
						URI:     "some-entry.tgz",
						SHA256:  dependencyHash,
						Version: "1.2.3",
					},
					"some-cnb-path",
					layerPath,
					"some-platform-dir",
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
			Expect(mappingResolver.FindDependencyMappingCall.Receives.PlatformDir).To(Equal("some-platform-dir"))

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

		context("when using the checksum field", func() {
			it.Before(func() {
				deliver = func() error {
					return service.Deliver(
						postal.Dependency{
							ID:       "some-entry",
							Stacks:   []string{"some-stack"},
							URI:      "some-entry.tgz",
							Checksum: fmt.Sprintf("sha512:%s", hash512),
							Version:  "1.2.3",
						},
						"some-cnb-path",
						layerPath,
						"some-platform-dir",
					)
				}
			})

			it("downloads the dependency and unpackages it into the path", func() {
				err := deliver()

				Expect(err).NotTo(HaveOccurred())

				Expect(transport.DropCall.Receives.Root).To(Equal("some-cnb-path"))
				Expect(transport.DropCall.Receives.Uri).To(Equal("some-entry.tgz"))
				Expect(mappingResolver.FindDependencyMappingCall.Receives.PlatformDir).To(Equal("some-platform-dir"))

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

		context("when the dependency has a strip-components value set", func() {
			it.Before(func() {
				var err error
				layerPath, err = os.MkdirTemp("", "path")
				Expect(err).NotTo(HaveOccurred())

				buffer := bytes.NewBuffer(nil)
				zw := gzip.NewWriter(buffer)
				tw := tar.NewWriter(zw)

				Expect(tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
				_, err = tw.Write(nil)
				Expect(err).NotTo(HaveOccurred())

				nestedFile := "some-dir/some-file"
				Expect(tw.WriteHeader(&tar.Header{Name: nestedFile, Mode: 0755, Size: int64(len(nestedFile))})).To(Succeed())
				_, err = tw.Write([]byte(nestedFile))
				Expect(err).NotTo(HaveOccurred())

				for _, file := range []string{"some-dir/first", "some-dir/second", "some-dir/third"} {
					Expect(tw.WriteHeader(&tar.Header{Name: file, Mode: 0755, Size: int64(len(file))})).To(Succeed())
					_, err = tw.Write([]byte(file))
					Expect(err).NotTo(HaveOccurred())
				}

				linkName := "some-dir/symlink"
				linkDest := "./first"
				Expect(tw.WriteHeader(&tar.Header{Name: linkName, Mode: 0777, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: linkDest})).To(Succeed())
				_, err = tw.Write([]byte{})
				Expect(err).NotTo(HaveOccurred())

				Expect(tw.Close()).To(Succeed())
				Expect(zw.Close()).To(Succeed())

				sum := sha256.Sum256(buffer.Bytes())
				dependencyHash = hex.EncodeToString(sum[:])

				transport.DropCall.Returns.ReadCloser = io.NopCloser(buffer)

				deliver = func() error {
					return service.Deliver(
						postal.Dependency{
							ID:              "some-entry",
							Stacks:          []string{"some-stack"},
							URI:             "some-entry.tgz",
							SHA256:          dependencyHash,
							Version:         "1.2.3",
							StripComponents: 1,
						},
						"some-cnb-path",
						layerPath,
						"",
					)
				}
			})

			it.After(func() {
				Expect(os.RemoveAll(layerPath)).To(Succeed())
			})

			it("downloads the dependency, strips given number of componenets and unpackages it into the path", func() {
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
					filepath.Join(layerPath, "symlink"),
					filepath.Join(layerPath, "some-file"),
				}))

				info, err := os.Stat(filepath.Join(layerPath, "first"))
				Expect(err).NotTo(HaveOccurred())
				Expect(info.Mode()).To(Equal(os.FileMode(0755)))
			})
		})

		context("when the dependency should be a named file", func() {
			it.Before(func() {
				var err error
				layerPath, err = os.MkdirTemp("", "path")
				Expect(err).NotTo(HaveOccurred())

				buffer := bytes.NewBuffer(nil)
				buffer.WriteString("some-file-contents")

				sum := sha256.Sum256(buffer.Bytes())
				dependencyHash = hex.EncodeToString(sum[:])

				transport.DropCall.Returns.ReadCloser = io.NopCloser(buffer)

				deliver = func() error {
					return service.Deliver(
						postal.Dependency{
							ID:      "some-entry",
							Stacks:  []string{"some-stack"},
							URI:     "https://dependencies.example.com/dependencies/some-file-name.txt",
							SHA256:  dependencyHash,
							Version: "1.2.3",
						},
						"some-cnb-path",
						layerPath,
						"some-platform-dir",
					)
				}
			})

			it.After(func() {
				Expect(os.RemoveAll(layerPath)).To(Succeed())
			})

			it("downloads the dependency and copies it into the path with the given name", func() {
				err := deliver()
				Expect(err).NotTo(HaveOccurred())

				Expect(transport.DropCall.Receives.Root).To(Equal("some-cnb-path"))
				Expect(transport.DropCall.Receives.Uri).To(Equal("https://dependencies.example.com/dependencies/some-file-name.txt"))

				files, err := filepath.Glob(fmt.Sprintf("%s/*", layerPath))
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(ConsistOf([]string{filepath.Join(layerPath, "some-file-name.txt")}))

				content, err := os.ReadFile(filepath.Join(layerPath, "some-file-name.txt"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("some-file-contents"))
			})
		})

		context("when there is a dependency mapping via binding", func() {
			it.Before(func() {
				mappingResolver.FindDependencyMappingCall.Returns.String = "dependency-mapping-entry.tgz"
			})

			context("the dependency has a checksum field", func() {
				it("looks up the dependency from the platform binding and downloads that instead", func() {
					err := deliver()

					Expect(err).NotTo(HaveOccurred())

					Expect(mappingResolver.FindDependencyMappingCall.Receives.Checksum).To(Equal("sha256:" + dependencyHash))
					Expect(mappingResolver.FindDependencyMappingCall.Receives.PlatformDir).To(Equal("some-platform-dir"))
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

			context("the dependency has a SHA256 field", func() {
				it("looks up the dependency from the platform binding and downloads that instead", func() {
					err := deliver()

					Expect(err).NotTo(HaveOccurred())

					Expect(mappingResolver.FindDependencyMappingCall.Receives.Checksum).To(Equal(fmt.Sprintf("sha256:%s", dependencyHash)))
					Expect(mappingResolver.FindDependencyMappingCall.Receives.PlatformDir).To(Equal("some-platform-dir"))
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
		})

		context("when there is a dependency mapping via binding", func() {
			it.Before(func() {
				mappingResolver.FindDependencyMappingCall.Returns.String = "dependency-mapping-entry.tgz"
			})

			it("looks up the dependency from the platform binding and downloads that instead", func() {
				err := deliver()

				Expect(err).NotTo(HaveOccurred())

				Expect(mappingResolver.FindDependencyMappingCall.Receives.Checksum).To(Equal(fmt.Sprintf("sha256:%s", dependencyHash)))
				Expect(mappingResolver.FindDependencyMappingCall.Receives.PlatformDir).To(Equal("some-platform-dir"))
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

		context("when there is a dependency mirror", func() {
			it.Before(func() {
				mirrorResolver.FindDependencyMirrorCall.Returns.String = "dependency-mirror-url"
			})

			it("downloads dependency from mirror", func() {
				err := deliver()

				Expect(err).NotTo(HaveOccurred())

				Expect(mirrorResolver.FindDependencyMirrorCall.Receives.Uri).To(Equal("some-entry.tgz"))
				Expect(mirrorResolver.FindDependencyMirrorCall.Receives.PlatformDir).To(Equal("some-platform-dir"))
				Expect(transport.DropCall.Receives.Root).To(Equal("some-cnb-path"))
				Expect(transport.DropCall.Receives.Uri).To(Equal("dependency-mirror-url"))

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
			context("when dependency mapping resolver fails", func() {
				it.Before(func() {
					mappingResolver.FindDependencyMappingCall.Returns.Error = fmt.Errorf("some dependency mapping error")
				})
				it("fails to find dependency mappings", func() {
					err := deliver()

					Expect(err).To(MatchError(ContainSubstring("some dependency mapping error")))
				})
			})

			context("when dependency mirror resolver fails", func() {
				it.Before(func() {
					mirrorResolver.FindDependencyMirrorCall.Returns.Error = fmt.Errorf("some dependency mirror error")
				})
				it("fails to find dependency mirror", func() {
					err := deliver()

					Expect(err).To(MatchError(ContainSubstring("some dependency mirror error")))
				})
			})

			context("when the transport cannot fetch a dependency", func() {
				it.Before(func() {
					transport.DropCall.Returns.Error = errors.New("there was an error")
				})

				it("returns an error", func() {
					err := deliver()

					Expect(err).To(MatchError("failed to fetch dependency: there was an error"))
				})
			})

			context("when there is a problem with the checksum", func() {
				it.Before(func() {
					deliver = func() error {
						return service.Deliver(
							postal.Dependency{
								ID:       "some-entry",
								Stacks:   []string{"some-stack"},
								URI:      "some-entry.tgz",
								Checksum: fmt.Sprintf("magic:%s", hash512),
								Version:  "1.2.3",
							},
							"some-cnb-path",
							layerPath,
							"some-platform-dir",
						)
					}
				})

				it("fails to create a validated reader", func() {
					err := deliver()

					Expect(err).To(MatchError(ContainSubstring(`unsupported algorithm "magic"`)))
				})
			})

			context("when the file contents are empty", func() {
				it.Before(func() {
					// This is a FLAC header
					buffer := bytes.NewBuffer([]byte("\x66\x4C\x61\x43\x00\x00\x00\x22"))
					transport.DropCall.Returns.ReadCloser = io.NopCloser(buffer)

					sum := sha256.Sum256(buffer.Bytes())
					dependencyHash = hex.EncodeToString(sum[:])
				})

				it("fails to create a gzip reader", func() {
					err := deliver()

					Expect(err).To(MatchError(ContainSubstring("unsupported archive type")))
				})
			})

			context("when the file checksum does not match", func() {
				it("fails to create a tar reader", func() {
					err := service.Deliver(
						postal.Dependency{
							ID:      "some-entry",
							Stacks:  []string{"some-stack"},
							URI:     "some-entry.tgz",
							SHA256:  "this is not a valid checksum",
							Version: "1.2.3",
						},
						"some-cnb-path",
						layerPath,
						"",
					)

					Expect(err).To(MatchError("validation error: checksum does not match"))
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

					linkName := "symlink"
					Expect(tw.WriteHeader(&tar.Header{Name: linkName, Mode: 0777, Size: int64(0), Typeflag: tar.TypeSymlink, Linkname: "some-file"})).To(Succeed())
					_, err := tw.Write([]byte{})
					Expect(err).NotTo(HaveOccurred())

					Expect(tw.Close()).To(Succeed())
					Expect(zw.Close()).To(Succeed())

					Expect(os.WriteFile(filepath.Join(layerPath, "some-file"), nil, 0644)).To(Succeed())
					Expect(os.Symlink("some-file", filepath.Join(layerPath, "symlink"))).To(Succeed())

					sum := sha256.Sum256(buffer.Bytes())
					dependencyHash = hex.EncodeToString(sum[:])

					transport.DropCall.Returns.ReadCloser = io.NopCloser(buffer)
				})

				it("fails to extract the symlink", func() {
					err := deliver()

					Expect(err).To(MatchError(ContainSubstring("failed to extract symlink")))
				})
			})

			context("when the has additional data in the byte stream", func() {
				it.Before(func() {
					var err error
					layerPath, err = os.MkdirTemp("", "path")
					Expect(err).NotTo(HaveOccurred())

					buffer := bytes.NewBuffer(nil)
					tw := tar.NewWriter(buffer)

					file := "some-file"
					Expect(tw.WriteHeader(&tar.Header{Name: file, Mode: 0755, Size: int64(len(file))})).To(Succeed())
					_, err = tw.Write([]byte(file))
					Expect(err).NotTo(HaveOccurred())

					Expect(tw.Close()).To(Succeed())

					sum := sha256.Sum256(buffer.Bytes())
					dependencyHash = hex.EncodeToString(sum[:])

					// Empty block is tricking tar reader into think that we have reached
					// EOF becuase we have surpassed the maximum block header size
					var block [1024]byte
					_, err = buffer.Write(block[:])
					Expect(err).NotTo(HaveOccurred())

					_, err = buffer.WriteString("additional data")
					Expect(err).NotTo(HaveOccurred())

					transport.DropCall.Returns.ReadCloser = io.NopCloser(buffer)

					deliver = func() error {
						return service.Deliver(
							postal.Dependency{
								ID:      "some-entry",
								Stacks:  []string{"some-stack"},
								URI:     "https://dependencies.example.com/dependencies/some-file-name.txt",
								SHA256:  dependencyHash,
								Version: "1.2.3",
							},
							"some-cnb-path",
							layerPath,
							"some-platform-dir",
						)
					}
				})

				it.After(func() {
					Expect(os.RemoveAll(layerPath)).To(Succeed())
				})

				it("returns an error", func() {
					err := deliver()
					Expect(err).To(MatchError("failed to validate dependency: checksum does not match"))

					Expect(transport.DropCall.Receives.Root).To(Equal("some-cnb-path"))
					Expect(transport.DropCall.Receives.Uri).To(Equal("https://dependencies.example.com/dependencies/some-file-name.txt"))
				})
			})
		})
	})

	context("GenerateBillOfMaterials", func() {
		it("returns a list of BOMEntry values", func() {
			entries := service.GenerateBillOfMaterials(
				postal.Dependency{
					ID:             "some-entry",
					Name:           "Some Entry",
					Checksum:       "sha256:some-sha",
					Source:         "some-source",
					SourceChecksum: "sha256:some-source-sha",
					Stacks:         []string{"some-stack"},
					URI:            "some-uri",
					Version:        "1.2.3",
				},
				postal.Dependency{
					ID:             "other-entry",
					Name:           "Other Entry",
					Checksum:       "sha256:other-sha",
					Source:         "other-source",
					SourceChecksum: "sha256:other-source-sha",
					Stacks:         []string{"other-stack"},
					URI:            "other-uri",
					Version:        "4.5.6",
				},
			)
			Expect(entries).To(Equal([]packit.BOMEntry{
				{
					Name: "Some Entry",
					Metadata: paketosbom.BOMMetadata{
						Checksum: paketosbom.BOMChecksum{
							Algorithm: paketosbom.SHA256,
							Hash:      "some-sha",
						},
						Source: paketosbom.BOMSource{
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "some-source-sha",
							},
							URI: "some-source",
						},

						URI:     "some-uri",
						Version: "1.2.3",
					},
				},
				{
					Name: "Other Entry",
					Metadata: paketosbom.BOMMetadata{
						Checksum: paketosbom.BOMChecksum{
							Algorithm: paketosbom.SHA256,
							Hash:      "other-sha",
						},
						Source: paketosbom.BOMSource{
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "other-source-sha",
							},
							URI: "other-source",
						},

						URI:     "other-uri",
						Version: "4.5.6",
					},
				},
			}))
		})

		context("when there is a CPE", func() {
			it("generates a BOM with the CPE", func() {
				entries := service.GenerateBillOfMaterials(
					postal.Dependency{
						CPE:            "some-cpe",
						ID:             "some-entry",
						Name:           "Some Entry",
						Checksum:       "sha256:some-sha",
						Source:         "some-source",
						SourceChecksum: "sha256:some-source-sha",
						Stacks:         []string{"some-stack"},
						URI:            "some-uri",
						Version:        "1.2.3",
					},
				)

				Expect(entries).To(Equal([]packit.BOMEntry{
					{
						Name: "Some Entry",
						Metadata: paketosbom.BOMMetadata{
							CPE: "some-cpe",
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "some-sha",
							},
							Source: paketosbom.BOMSource{
								Checksum: paketosbom.BOMChecksum{
									Algorithm: paketosbom.SHA256,
									Hash:      "some-source-sha",
								},
								URI: "some-source",
							},

							URI:     "some-uri",
							Version: "1.2.3",
						},
					},
				}))
			})
			context("and there are CPEs", func() {
				it("uses CPE, ignores CPEs, for backward compatibility", func() {
					entries := service.GenerateBillOfMaterials(
						postal.Dependency{
							CPE:            "some-cpe",
							CPEs:           []string{"some-other-cpe"},
							ID:             "some-entry",
							Name:           "Some Entry",
							Checksum:       "sha256:some-sha",
							Source:         "some-source",
							SourceChecksum: "sha256:some-source-sha",
							Stacks:         []string{"some-stack"},
							URI:            "some-uri",
							Version:        "1.2.3",
						},
					)

					Expect(entries).To(Equal([]packit.BOMEntry{
						{
							Name: "Some Entry",
							Metadata: paketosbom.BOMMetadata{
								CPE: "some-cpe",
								Checksum: paketosbom.BOMChecksum{
									Algorithm: paketosbom.SHA256,
									Hash:      "some-sha",
								},
								Source: paketosbom.BOMSource{
									Checksum: paketosbom.BOMChecksum{
										Algorithm: paketosbom.SHA256,
										Hash:      "some-source-sha",
									},
									URI: "some-source",
								},

								URI:     "some-uri",
								Version: "1.2.3",
							},
						},
					}))
				})

			})
		})

		context("when there is a deprecation date", func() {
			var deprecationDate time.Time

			it.Before(func() {
				var err error
				deprecationDate, err = time.Parse(time.RFC3339, "2022-04-01T00:00:00Z")
				Expect(err).NotTo(HaveOccurred())
			})

			it("generates a BOM with the deprecation date", func() {
				entries := service.GenerateBillOfMaterials(
					postal.Dependency{
						DeprecationDate: deprecationDate,
						ID:              "some-entry",
						Name:            "Some Entry",
						Checksum:        "sha256:some-sha",
						Source:          "some-source",
						SourceChecksum:  "sha256:some-source-sha",
						Stacks:          []string{"some-stack"},
						URI:             "some-uri",
						Version:         "1.2.3",
					},
				)

				Expect(entries).To(Equal([]packit.BOMEntry{
					{
						Name: "Some Entry",
						Metadata: paketosbom.BOMMetadata{
							DeprecationDate: deprecationDate,
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "some-sha",
							},
							Source: paketosbom.BOMSource{
								Checksum: paketosbom.BOMChecksum{
									Algorithm: paketosbom.SHA256,
									Hash:      "some-source-sha",
								},
								URI: "some-source",
							},

							URI:     "some-uri",
							Version: "1.2.3",
						},
					},
				}))
			})
		})

		context("when there is license information", func() {
			it("generates a BOM with the license information", func() {
				entries := service.GenerateBillOfMaterials(
					postal.Dependency{
						ID:             "some-entry",
						Licenses:       []string{"some-license"},
						Name:           "Some Entry",
						Checksum:       "sha256:some-sha",
						Source:         "some-source",
						SourceChecksum: "sha256:some-source-sha",
						Stacks:         []string{"some-stack"},
						URI:            "some-uri",
						Version:        "1.2.3",
					},
				)

				Expect(entries).To(Equal([]packit.BOMEntry{
					{
						Name: "Some Entry",
						Metadata: paketosbom.BOMMetadata{
							Licenses: []string{"some-license"},
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "some-sha",
							},
							Source: paketosbom.BOMSource{
								Checksum: paketosbom.BOMChecksum{
									Algorithm: paketosbom.SHA256,
									Hash:      "some-source-sha",
								},
								URI: "some-source",
							},

							URI:     "some-uri",
							Version: "1.2.3",
						},
					},
				}))
			})
		})

		context("when there is a pURL", func() {
			it("generates a BOM with the pURL", func() {
				entries := service.GenerateBillOfMaterials(
					postal.Dependency{
						ID:             "some-entry",
						Name:           "Some Entry",
						PURL:           "some-purl",
						Checksum:       "sha256:some-sha",
						Source:         "some-source",
						SourceChecksum: "sha256:some-source-sha",
						Stacks:         []string{"some-stack"},
						URI:            "some-uri",
						Version:        "1.2.3",
					},
				)

				Expect(entries).To(Equal([]packit.BOMEntry{
					{
						Name: "Some Entry",
						Metadata: paketosbom.BOMMetadata{
							PURL: "some-purl",
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "some-sha",
							},
							Source: paketosbom.BOMSource{
								Checksum: paketosbom.BOMChecksum{
									Algorithm: paketosbom.SHA256,
									Hash:      "some-source-sha",
								},
								URI: "some-source",
							},

							URI:     "some-uri",
							Version: "1.2.3",
						},
					},
				}))
			})
		})

		context("when there is a SHA256 instead of Checksum", func() {
			it("generates a BOM with the SHA256", func() {
				entries := service.GenerateBillOfMaterials(
					postal.Dependency{
						ID:           "some-entry",
						Name:         "Some Entry",
						SHA256:       "some-sha",
						Source:       "some-source",
						SourceSHA256: "some-source-sha",
						Stacks:       []string{"some-stack"},
						URI:          "some-uri",
						Version:      "1.2.3",
					},
				)

				Expect(entries).To(Equal([]packit.BOMEntry{
					{
						Name: "Some Entry",
						Metadata: paketosbom.BOMMetadata{
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "some-sha",
							},
							Source: paketosbom.BOMSource{
								Checksum: paketosbom.BOMChecksum{
									Algorithm: paketosbom.SHA256,
									Hash:      "some-source-sha",
								},
								URI: "some-source",
							},

							URI:     "some-uri",
							Version: "1.2.3",
						},
					},
				}))
			})
		})

		context("when there is a checksum and SHA256", func() {
			it("generates a BOM with the checksum", func() {
				entries := service.GenerateBillOfMaterials(
					postal.Dependency{
						ID:             "some-entry",
						Name:           "Some Entry",
						Checksum:       "sha512:checksum-sha",
						SHA256:         "some-sha",
						Source:         "some-source",
						SourceChecksum: "sha512:source-checksum-sha",
						SourceSHA256:   "some-source-sha",
						Stacks:         []string{"some-stack"},
						URI:            "some-uri",
						Version:        "1.2.3",
					},
				)

				Expect(entries).To(Equal([]packit.BOMEntry{
					{
						Name: "Some Entry",
						Metadata: paketosbom.BOMMetadata{
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA512,
								Hash:      "checksum-sha",
							},
							Source: paketosbom.BOMSource{
								Checksum: paketosbom.BOMChecksum{
									Algorithm: paketosbom.SHA512,
									Hash:      "source-checksum-sha",
								},
								URI: "some-source",
							},

							URI:     "some-uri",
							Version: "1.2.3",
						},
					},
				}))
			})
		})

		context("when the checksum algorithm is unknown", func() {
			it("generates a BOM with the empty/unknown checksum", func() {
				entries := service.GenerateBillOfMaterials(
					postal.Dependency{
						ID:             "some-entry",
						Name:           "Some Entry",
						Checksum:       "no-such-algo:some-hash",
						Source:         "some-source",
						SourceChecksum: "no-such-algo:some-hash",
						Stacks:         []string{"some-stack"},
						URI:            "some-uri",
						Version:        "1.2.3",
					},
				)

				Expect(entries).To(Equal([]packit.BOMEntry{
					{
						Name: "Some Entry",
						Metadata: paketosbom.BOMMetadata{
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.UNKNOWN,
								Hash:      "",
							},
							Source: paketosbom.BOMSource{
								Checksum: paketosbom.BOMChecksum{
									Algorithm: paketosbom.UNKNOWN,
									Hash:      "",
								},
								URI: "some-source",
							},

							URI:     "some-uri",
							Version: "1.2.3",
						},
					},
				}))
			})
		})

		context("when there is no checksum or SHA256", func() {
			it("generates a BOM with the empty/unknown checksum", func() {
				entries := service.GenerateBillOfMaterials(
					postal.Dependency{
						ID:      "some-entry",
						Name:    "Some Entry",
						Source:  "some-source",
						Stacks:  []string{"some-stack"},
						URI:     "some-uri",
						Version: "1.2.3",
					},
				)

				Expect(entries).To(Equal([]packit.BOMEntry{
					{
						Name: "Some Entry",
						Metadata: paketosbom.BOMMetadata{
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.UNKNOWN,
								Hash:      "",
							},
							Source: paketosbom.BOMSource{
								Checksum: paketosbom.BOMChecksum{
									Algorithm: paketosbom.UNKNOWN,
									Hash:      "",
								},
								URI: "some-source",
							},

							URI:     "some-uri",
							Version: "1.2.3",
						},
					},
				}))
			})
		})
	})
}
