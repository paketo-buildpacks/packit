package commands_test

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitCommands(t *testing.T) {
	suite := spec.New("jam/commands", spec.Report(report.Terminal{}))
	suite("Pack", testPack)
	suite.Run(t)
}

func ExtractFile(file *os.File, name string) ([]byte, error) {
	_, err := file.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	//TODO: Replace me with decompression library
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if hdr.Name == name {
			return ioutil.ReadAll(tr)
		}
	}

	return nil, fmt.Errorf("no such file: %s", name)
}
