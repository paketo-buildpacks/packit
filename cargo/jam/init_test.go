package main_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var path string

func TestUnitJam(t *testing.T) {
	suite := spec.New("cargo/jam", spec.Report(report.Terminal{}))
	suite("pack", testPack)
	suite("Errors", testErrors)

	suite.Before(func(t *testing.T) {
		var (
			Expect = NewWithT(t).Expect
			err    error
		)

		path, err = gexec.Build("github.com/cloudfoundry/packit/cargo/jam")
		Expect(err).NotTo(HaveOccurred())
	})

	suite.After(func(t *testing.T) {
		gexec.CleanupBuildArtifacts()
	})

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

type Buffer struct {
	b bytes.Buffer
	m sync.Mutex
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Read(p)
}
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}
func (b *Buffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}
