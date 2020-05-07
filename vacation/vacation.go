package vacation

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

type TarArchive struct {
	reader     io.Reader
	components int
}

type TarGzipArchive struct {
	reader     io.Reader
	components int
}

type TarXZArchive struct {
	reader     io.Reader
	components int
}

func NewTarArchive(inputReader io.Reader) TarArchive {
	return TarArchive{reader: inputReader}
}

func NewTarGzipArchive(inputReader io.Reader) TarGzipArchive {
	return TarGzipArchive{reader: inputReader}
}

func NewTarXZArchive(inputReader io.Reader) TarXZArchive {
	return TarXZArchive{reader: inputReader}
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

		fileNames := strings.Split(hdr.Name, string(filepath.Separator))

		if len(fileNames) <= ta.components {
			continue
		}

		path := filepath.Join(append([]string{destination}, fileNames[ta.components:]...)...)
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
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}

	return NewTarArchive(gzr).StripComponents(gz.components).Decompress(destination)
}

func (txz TarXZArchive) Decompress(destination string) error {
	xzr, err := xz.NewReader(txz.reader)
	if err != nil {
		return fmt.Errorf("failed to create xz reader: %w", err)
	}

	return NewTarArchive(xzr).StripComponents(txz.components).Decompress(destination)
}

func (ta TarArchive) StripComponents(components int) TarArchive {
	ta.components = components
	return ta
}

func (gz TarGzipArchive) StripComponents(components int) TarGzipArchive {
	gz.components = components
	return gz
}

func (txz TarXZArchive) StripComponents(components int) TarXZArchive {
	txz.components = components
	return txz
}

type ZipArchive struct {
	reader io.Reader
}

func NewZipArchive(inputReader io.Reader) ZipArchive {
	return ZipArchive{reader: inputReader}
}

func (z ZipArchive) Decompress(destination string) error {
	// Have to convert and io.Reader into a bytes.Reader which
	// implements the ReadAt function making it compatible with
	// the io.ReaderAt inteface which required for zip.NewReader
	buff := bytes.NewBuffer(nil)
	size, err := io.Copy(buff, z.reader)
	if err != nil {
		return err
	}

	readerAt := bytes.NewReader(buff.Bytes())

	zr, err := zip.NewReader(readerAt, size)
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	for _, f := range zr.File {
		path := filepath.Join(destination, f.Name)

		switch {
		case f.FileInfo().IsDir():
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to unzip directory: %w", err)
			}
		case f.FileInfo().Mode()&os.ModeSymlink != 0:
			fd, err := f.Open()
			if err != nil {
				return err
			}

			content, err := ioutil.ReadAll(fd)
			if err != nil {
				return err
			}

			err = os.Symlink(string(content), path)
			if err != nil {
				return fmt.Errorf("failed to unzip symlink: %w", err)
			}
		default:
			err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to unzip directory that was part of file path: %w", err)
			}

			dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return fmt.Errorf("failed to unzip file: %w", err)
			}
			defer dst.Close()

			src, err := f.Open()
			if err != nil {
				return err
			}
			defer src.Close()

			_, err = io.Copy(dst, src)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
