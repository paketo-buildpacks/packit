package postal_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudfoundry/packit/postal"
	"github.com/cloudfoundry/packit/postal/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testService(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path      string
		transport *fakes.Transport
		service   postal.Service
	)

	it.Before(func() {
		file, err := ioutil.TempFile("", "buildpack.toml")
		Expect(err).NotTo(HaveOccurred())

		path = file.Name()

		_, err = file.WriteString(`
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
`)
		Expect(err).NotTo(HaveOccurred())

		Expect(file.Close()).To(Succeed())

		transport = &fakes.Transport{}

		service = postal.NewService(transport)
	})

	context("Resolve", func() {
		it("finds the best matching dependency given a plan entry", func() {
			dependency, err := service.Resolve(path, "some-entry", "1.2.*", "some-stack")
			Expect(err).NotTo(HaveOccurred())
			Expect(dependency).To(Equal(postal.Dependency{
				ID:      "some-entry",
				Stacks:  []string{"some-stack"},
				URI:     "some-uri",
				SHA256:  "some-sha",
				Version: "1.2.3",
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
		})

		context("when there is a default version", func() {
			it.Before(func() {
				err := ioutil.WriteFile(path, []byte(`
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
`), 0644)
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
					err := ioutil.WriteFile(path, []byte("this is not toml"), 0644)
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
					err := ioutil.WriteFile(path, []byte(`
[[metadata.dependencies]]
id = "some-entry"
sha256 = "some-sha"
stacks = ["some-stack"]
uri = "some-uri"
version = "this is super not semver"
`), 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				it("returns an error", func() {
					_, err := service.Resolve(path, "some-entry", "1.2.3", "some-stack")
					Expect(err).To(MatchError(ContainSubstring("Invalid Semantic Version")))
				})
			})

			context("when the entry version constraint cannot be satisfied", func() {
				it("return an error", func() {
					_, err := service.Resolve(path, "some-entry", "9.9.9", "some-stack")
					Expect(err).To(MatchError(ContainSubstring("failed to satisfy \"some-entry\" dependency version constraint \"9.9.9\": no compatible versions")))
				})
			})
		})
	})

	context("Install", func() {
		var (
			dependencySHA string
			tmpDir        string
		)

		it.Before(func() {
			var err error
			tmpDir, err = ioutil.TempDir("", "path")
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

			transport.DropCall.Returns.ReadCloser = ioutil.NopCloser(buffer)
		})

		it.After(func() {
			Expect(os.RemoveAll(tmpDir)).To(Succeed())
		})

		it("downloads the dependency and unpackages it into the path", func() {
			err := service.Install(postal.Dependency{
				ID:      "some-entry",
				Stacks:  []string{"some-stack"},
				URI:     "some-entry.tgz",
				SHA256:  dependencySHA,
				Version: "1.2.3",
			}, "some-cnb-path", tmpDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(transport.DropCall.Receives.Root).To(Equal("some-cnb-path"))
			Expect(transport.DropCall.Receives.Uri).To(Equal("some-entry.tgz"))

			files, err := filepath.Glob(fmt.Sprintf("%s/*", tmpDir))
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(ConsistOf([]string{
				filepath.Join(tmpDir, "first"),
				filepath.Join(tmpDir, "second"),
				filepath.Join(tmpDir, "third"),
				filepath.Join(tmpDir, "some-dir"),
				filepath.Join(tmpDir, "symlink"),
			}))

			info, err := os.Stat(filepath.Join(tmpDir, "first"))
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode()).To(Equal(os.FileMode(0755)))
		})

		context("failure cases", func() {
			context("when the transport cannot fetch a dependency", func() {
				it.Before(func() {
					transport.DropCall.Returns.Error = errors.New("there was an error")
				})

				it("returns an error", func() {
					err := service.Install(postal.Dependency{
						ID:      "some-entry",
						Stacks:  []string{"some-stack"},
						URI:     "some-entry.tgz",
						SHA256:  dependencySHA,
						Version: "1.2.3",
					}, "some-cnb-path", tmpDir)
					Expect(err).To(MatchError("failed to fetch dependency: there was an error"))
				})
			})

			context("when the file contents are empty", func() {
				it.Before(func() {
					transport.DropCall.Returns.ReadCloser = ioutil.NopCloser(strings.NewReader(""))
				})

				it("fails to create a gzip reader", func() {
					err := service.Install(postal.Dependency{
						ID:      "some-entry",
						Stacks:  []string{"some-stack"},
						URI:     "some-entry.tgz",
						SHA256:  dependencySHA,
						Version: "1.2.3",
					}, "some-cnb-path", tmpDir)

					Expect(err).To(MatchError(ContainSubstring("failed to create gzip reader")))
				})
			})

			context("when the file contents are malformed", func() {
				it.Before(func() {
					buffer := bytes.NewBuffer(nil)
					gzipWriter := gzip.NewWriter(buffer)

					_, err := gzipWriter.Write([]byte("something"))
					Expect(err).NotTo(HaveOccurred())

					Expect(gzipWriter.Close()).To(Succeed())

					transport.DropCall.Returns.ReadCloser = ioutil.NopCloser(buffer)
				})

				it("fails to create a tar reader", func() {
					err := service.Install(postal.Dependency{
						ID:      "some-entry",
						Stacks:  []string{"some-stack"},
						URI:     "some-entry.tgz",
						SHA256:  dependencySHA,
						Version: "1.2.3",
					}, "some-cnb-path", tmpDir)

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
					}, "some-cnb-path", tmpDir)

					Expect(err).To(MatchError(ContainSubstring("checksum does not match")))
				})
			})

			context("when it does not have permission to write into directory on container", func() {
				it.Before(func() {
					Expect(os.Chmod(tmpDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(tmpDir, 0755)).To(Succeed())
				})

				it("fails to make a dir", func() {
					err := service.Install(postal.Dependency{
						ID:      "some-entry",
						Stacks:  []string{"some-stack"},
						URI:     "some-entry.tgz",
						SHA256:  dependencySHA,
						Version: "1.2.3",
					}, "some-cnb-path", tmpDir)

					Expect(err).To(MatchError(ContainSubstring("failed to create archived directory")))
				})
			})

			context("when it does not have permission to write into directory that it decompressed", func() {
				var testDir string
				it.Before(func() {
					testDir = filepath.Join(tmpDir, "some-dir")
					Expect(os.MkdirAll(testDir, os.ModePerm)).To(Succeed())
					Expect(os.Chmod(testDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(testDir, 0755)).To(Succeed())
				})

				it("fails to make a file", func() {
					err := service.Install(postal.Dependency{
						ID:      "some-entry",
						Stacks:  []string{"some-stack"},
						URI:     "some-entry.tgz",
						SHA256:  dependencySHA,
						Version: "1.2.3",
					}, "some-cnb-path", tmpDir)

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

					transport.DropCall.Returns.ReadCloser = ioutil.NopCloser(buffer)
				})

				it("fails to extract the symlink", func() {
					err := service.Install(postal.Dependency{
						ID:      "some-entry",
						Stacks:  []string{"some-stack"},
						URI:     "some-entry.tgz",
						SHA256:  dependencySHA,
						Version: "1.2.3",
					}, "some-cnb-path", tmpDir)
					Expect(err).To(MatchError(ContainSubstring("failed to extract symlink")))
				})
			})
		})
	})
}
