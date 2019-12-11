package vacation

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type TarReadyReader struct {
	io.Reader
}

func NewTarReader(inputReader io.Reader) (TarReadyReader, error) {
	return TarReadyReader{inputReader}, nil
}

func NewGzipTarReader(inputReader io.Reader) (TarReadyReader, error) {
	gzipReader, err := gzip.NewReader(inputReader)
	if err != nil {
		return TarReadyReader{nil}, fmt.Errorf("failed to create gzip reader: %s", err.Error())
	}
	return TarReadyReader{gzipReader}, nil
}

func (tr TarReadyReader) Decompress(destination string) error {
	tarReader := tar.NewReader(tr)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar response: %s", err.Error())
		}

		path := filepath.Join(destination, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(path, hdr.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("failed to create archived directory: %s", err.Error())
			}

		case tar.TypeReg:
			file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, hdr.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("failed to create archived file %s", err.Error())
			}

			_, err = io.Copy(file, tarReader)
			if err != nil {
				return err
			}

			err = file.Close()
			if err != nil {
				return err
			}
		}
	}

	return nil
}
