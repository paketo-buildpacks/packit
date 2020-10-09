// Package vacation provides a set of functions that enable input stream
// decompression logic from several popular decompression formats. This allows
// from decompression from either a file or any other byte stream, which is
// useful for decompressing files that are being downloaded.
package vacation

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/ulikunitz/xz"
)

// An Archive decompresses tar, gzip and xz compressed tar, and zip files from
// an input stream.
type Archive struct {
	reader     io.Reader
	components int
}

// A TarArchive decompresses tar files from an input stream.
type TarArchive struct {
	reader     io.Reader
	components int
}

// A TarGzipArchive decompresses gziped tar files from an input stream.
type TarGzipArchive struct {
	reader     io.Reader
	components int
}

// A TarXZArchive decompresses xz tar files from an input stream.
type TarXZArchive struct {
	reader     io.Reader
	components int
}

// NewArchive returns a new Archive that reads from inputReader.
func NewArchive(inputReader io.Reader) Archive {
	return Archive{reader: inputReader}
}

// NewTarArchive returns a new TarArchive that reads from inputReader.
func NewTarArchive(inputReader io.Reader) TarArchive {
	return TarArchive{reader: inputReader}
}

// NewTarGzipArchive returns a new TarGzipArchive that reads from inputReader.
func NewTarGzipArchive(inputReader io.Reader) TarGzipArchive {
	return TarGzipArchive{reader: inputReader}
}

// NewTarXZArchive returns a new TarXZArchive that reads from inputReader.
func NewTarXZArchive(inputReader io.Reader) TarXZArchive {
	return TarXZArchive{reader: inputReader}
}

// Decompress reads from TarArchive and writes files into the
// destination specified.
func (ta TarArchive) Decompress(destination string) error {
	// This map keeps track of what directories have been made already so that we
	// only attempt to make them once for a cleaner interaction.  This map is
	// only necessary in cases where there are no directory headers in the
	// tarball, which can be seen in the test around there being no directory
	// metadata.
	directories := map[string]interface{}{}

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

		// Checks to see if file should be written when stripping components
		if len(fileNames) <= ta.components {
			continue
		}

		// Constructs the path that conforms to the stripped components.
		path := filepath.Join(append([]string{destination}, fileNames[ta.components:]...)...)

		// This switch case handles all cases for creating the directory structure
		// this logic is needed to handle tarballs with no directory headers.
		switch hdr.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create archived directory: %s", err)
			}

			directories[path] = nil

		default:
			dir := filepath.Dir(path)
			_, ok := directories[dir]
			if !ok {
				err = os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					return fmt.Errorf("failed to create archived directory from file path: %s", err)
				}
				directories[dir] = nil
			}
		}

		// This switch case handles the creation of files during the untaring process.
		switch hdr.Typeflag {
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

// Decompress reads from Archive, determines the archive type of the input
// stream, and writes files into the destination specified.
func (a Archive) Decompress(destination string) error {
	// Convert reader into a buffered read so that the header can be peeked to
	// determine the type.
	bufferedReader := bufio.NewReader(a.reader)

	// The number 3072 is lifted from the mimetype library and the definition of
	// the constant at the time of writing this functionality is listed below.
	// https://github.com/gabriel-vasile/mimetype/blob/c64c025a7c2d8d45ba57d3cebb50a1dbedb3ed7e/internal/matchers/matchers.go#L6
	header, err := bufferedReader.Peek(3072)
	if err != nil && err != io.EOF {
		return err
	}

	mime := mimetype.Detect(header)

	// This switch case is reponsible for determining what the decompression
	// startegy should be.
	switch mime.String() {
	case "application/x-tar":
		return NewTarArchive(bufferedReader).StripComponents(a.components).Decompress(destination)
	case "application/gzip":
		return NewTarGzipArchive(bufferedReader).StripComponents(a.components).Decompress(destination)
	case "application/x-xz":
		return NewTarXZArchive(bufferedReader).StripComponents(a.components).Decompress(destination)
	case "application/zip":
		return NewZipArchive(bufferedReader).Decompress(destination)
	default:
		return fmt.Errorf("unsupported archive type: %s", mime.String())
	}
}

// Decompress reads from TarGzipArchive and writes files into the destination
// specified.
func (gz TarGzipArchive) Decompress(destination string) error {
	gzr, err := gzip.NewReader(gz.reader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}

	return NewTarArchive(gzr).StripComponents(gz.components).Decompress(destination)
}

// Decompress reads from TarXZArchive and writes files into the destination
// specified.
func (txz TarXZArchive) Decompress(destination string) error {
	xzr, err := xz.NewReader(txz.reader)
	if err != nil {
		return fmt.Errorf("failed to create xz reader: %w", err)
	}

	return NewTarArchive(xzr).StripComponents(txz.components).Decompress(destination)
}

// StripComponents behaves like the --strip-components flag on tar command
// removing the first n levels from the final decompression destination.
// Setting this is a no-op for archive types that do not use --strip-components
// (such as zip).
func (a Archive) StripComponents(components int) Archive {
	a.components = components
	return a
}

// StripComponents behaves like the --strip-components flag on tar command
// removing the first n levels from the final decompression destination.
func (ta TarArchive) StripComponents(components int) TarArchive {
	ta.components = components
	return ta
}

// StripComponents behaves like the --strip-components flag on tar command
// removing the first n levels from the final decompression destination.
func (gz TarGzipArchive) StripComponents(components int) TarGzipArchive {
	gz.components = components
	return gz
}

// StripComponents behaves like the --strip-components flag on tar command
// removing the first n levels from the final decompression destination.
func (txz TarXZArchive) StripComponents(components int) TarXZArchive {
	txz.components = components
	return txz
}

// A ZipArchive decompresses zip files from an input stream.
type ZipArchive struct {
	reader io.Reader
}

// NewZipArchive returns a new ZipArchive that reads from inputReader.
func NewZipArchive(inputReader io.Reader) ZipArchive {
	return ZipArchive{reader: inputReader}
}

// Decompress reads from ZipArchive and writes files into the destination
// specified.
func (z ZipArchive) Decompress(destination string) error {
	// Have to convert an io.Reader into a bytes.Reader which implements the
	// ReadAt function making it compatible with the io.ReaderAt inteface which
	// required for zip.NewReader
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
