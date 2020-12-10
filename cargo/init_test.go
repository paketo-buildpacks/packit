package cargo_test

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitCargo(t *testing.T) {
	suite := spec.New("cargo", spec.Report(report.Terminal{}))
	suite("BuildpackParser", testBuildpackParser)
	suite("Config", testConfig)
	suite("DependencyCacher", testDependencyCacher)
	suite("DirectoryDuplicator", testDirectoryDuplicator)
	suite("FileBundler", testFileBundler)
	suite("PrePackager", testPrePackager)
	suite("TarBuilder", testTarBuilder)
	suite("Transport", testTransport)
	suite("ValidatedReader", testValidatedReader)
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
			contents, err := ioutil.ReadAll(tr)
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
