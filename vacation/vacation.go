package vacation

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type TarArchive struct {
	reader io.Reader
}

type TarGzipArchive struct {
	reader io.Reader
}

func NewTarArchive(inputReader io.Reader) TarArchive {
	return TarArchive{reader: inputReader}
}

func NewTarGzipArchive(inputReader io.Reader) TarGzipArchive {
	return TarGzipArchive{reader: inputReader}
}

func (ta TarArchive) Decompress(destination string) error {
	tarReader := tar.NewReader(ta.reader)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar response: %s", err)
		}

		path := filepath.Join(destination, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create archived directory: %s", err)
			}

		case tar.TypeReg:
			file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, hdr.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("failed to create archived file: %s", err)
			}

			_, err = io.Copy(file, tarReader)
			if err != nil {
				return err
			}

			err = file.Close()
			if err != nil {
				return err
			}

		case tar.TypeSymlink:
			err = os.Symlink(hdr.Linkname, path)
			if err != nil {
				return fmt.Errorf("failed to extract symlink: %s", err)
			}

		}
	}

	return nil
}

func (gz TarGzipArchive) Decompress(destination string) error {
	gzr, err := gzip.NewReader(gz.reader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %s", err)
	}

	return NewTarArchive(gzr).Decompress(destination)
}
