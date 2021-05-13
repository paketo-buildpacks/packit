package internal_test

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitCargo(t *testing.T) {
	gomega.SetDefaultEventuallyTimeout(10 * time.Second)

	suite := spec.New("cargo/jam/internal", spec.Report(report.Terminal{}))
	suite("BuilderConfig", testBuilderConfig)
	suite("BuildpackConfig", testBuildpackConfig)
	suite("BuildpackInspector", testBuildpackInspector)
	suite("DependencyCacher", testDependencyCacher)
	suite("Dependency", testDependency)
	suite("FileBundler", testFileBundler)
	suite("Formatter", testFormatter)
	suite("Image", testImage)
	suite("PrePackager", testPrePackager)
	suite("PackageConfig", testPackageConfig)
	suite("TarBuilder", testTarBuilder)
	suite.Run(t)
}

func ExtractFile(file *os.File, name string) ([]byte, *tar.Header, error) {
	_, err := file.Seek(0, 0)
	if err != nil {
		return nil, nil, err
	}

	//TODO: Replace me with decompression library
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return nil, nil, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, nil, err
		}

		if hdr.Name == name {
			contents, err := io.ReadAll(tr)
			if err != nil {
				return nil, nil, err
			}

			return contents, hdr, nil
		}
	}

	return nil, nil, fmt.Errorf("no such file: %s", name)
}

type errorReader struct{}

func (r errorReader) Read(p []byte) (int, error) {
	return 0, errors.New("failed to read")
}
