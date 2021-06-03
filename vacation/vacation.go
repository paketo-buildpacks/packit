// Package vacation provides a set of functions that enable input stream
// decompression logic from several popular decompression formats. This allows
// from decompression from either a file or any other byte stream, which is
// useful for decompressing files that are being downloaded.
package vacation

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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

	// Struct and slice to collect symlinks and create them after all files have
	// been created
	type header struct {
		name     string
		linkname string
		path     string
	}

	var symlinkHeaders []header

	tarReader := tar.NewReader(ta.reader)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar response: %s", err)
		}

		// Skip if the destination it the destination directory itself i.e. ./
		if hdr.Name == "./" {
			continue
		}

		err = checkExtractPath(hdr.Name, destination)
		if err != nil {
			return err
		}

		fileNames := strings.Split(hdr.Name, "/")

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
			// Collect all of the headers for symlinks so that they can be verified
			// after all other files are written
			symlinkHeaders = append(symlinkHeaders, header{
				name:     hdr.Name,
				linkname: hdr.Linkname,
				path:     path,
			})
		}
	}

	// Sort the symlinks so that symlinks of symlinks have their base link
	// created before they are created.
	//
	// For example:
	// b-sym -> a-sym/x
	// a-sym -> z
	// c-sym -> d-sym
	// d-sym -> z
	//
	// Will sort to:
	// a-sym -> z
	// b-sym -> a-sym/x
	// d-sym -> z
	// c-sym -> d-sym
	sort.Slice(symlinkHeaders, func(i, j int) bool {
		if filepath.Clean(symlinkHeaders[i].name) == linknameFullPath(symlinkHeaders[j].name, symlinkHeaders[j].linkname) {
			return true
		}

		if filepath.Clean(symlinkHeaders[j].name) == linknameFullPath(symlinkHeaders[i].name, symlinkHeaders[i].linkname) {
			return false
		}

		return filepath.Clean(symlinkHeaders[i].name) < linknameFullPath(symlinkHeaders[j].name, symlinkHeaders[j].linkname)
	})

	for _, h := range symlinkHeaders {
		evalPath := linknameFullPath(h.path, h.linkname)
		// Don't use constucted link if the link is absolute
		if filepath.IsAbs(h.linkname) {
			evalPath = h.linkname
		}

		// Check to see if the file that will be linked to is valid for symlinking
		_, err := filepath.EvalSymlinks(evalPath)
		if err != nil {
			return fmt.Errorf("failed to evaluate symlink %s: %w", h.path, err)
		}

		err = os.Symlink(h.linkname, h.path)
		if err != nil {
			return fmt.Errorf("failed to extract symlink: %s", err)
		}
	}

	return nil
}

// Decompress reads from Archive, determines the archive type of the input
// stream, and writes files into the destination specified.
//
// Archive decompression will also handle files that are types "text/plain;
// charset=utf-8" and write the contents of the input stream to a file name
// "artifact" in the destination directory.
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
	case "text/plain; charset=utf-8":
		// This function will write the contents of the reader to file called
		// "artifact" in the destination directory
		return writeTextFile(bufferedReader, destination)
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

func writeTextFile(reader io.Reader, destination string) error {
	file, err := os.Create(filepath.Join(destination, "artifact"))
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}

	return nil
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
	// Struct and slice to collect symlinks and create them after all files have
	// been created
	type header struct {
		name     string
		linkname string
		path     string
	}

	var symlinkHeaders []header

	// Use an os.File to buffer the zip contents. This is needed because
	// zip.NewReader requires an io.ReaderAt so that it can jump around within
	// the file as it decompresses.
	buffer, err := os.CreateTemp("", "")
	if err != nil {
		return err
	}
	defer os.Remove(buffer.Name())

	size, err := io.Copy(buffer, z.reader)
	if err != nil {
		return err
	}

	zr, err := zip.NewReader(buffer, size)
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	for _, f := range zr.File {
		// Skip if the destination it the destination directory itself i.e. ./
		if f.Name == "./" {
			continue
		}

		err = checkExtractPath(f.Name, destination)
		if err != nil {
			return err
		}

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

			linkname, err := io.ReadAll(fd)
			if err != nil {
				return err
			}

			// Collect all of the headers for symlinks so that they can be verified
			// after all other files are written
			symlinkHeaders = append(symlinkHeaders, header{
				name:     f.Name,
				linkname: string(linkname),
				path:     path,
			})

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

	// Sort the symlinks so that symlinks of symlinks have their base link
	// created before they are created.
	//
	// For example:
	// b-sym -> a-sym/x
	// a-sym -> z
	// c-sym -> d-sym
	// d-sym -> z
	//
	// Will sort to:
	// a-sym -> z
	// b-sym -> a-sym/x
	// d-sym -> z
	// c-sym -> d-sym
	sort.Slice(symlinkHeaders, func(i, j int) bool {
		if filepath.Clean(symlinkHeaders[i].name) == linknameFullPath(symlinkHeaders[j].name, symlinkHeaders[j].linkname) {
			return true
		}

		if filepath.Clean(symlinkHeaders[j].name) == linknameFullPath(symlinkHeaders[i].name, symlinkHeaders[i].linkname) {
			return false
		}

		return filepath.Clean(symlinkHeaders[i].name) < linknameFullPath(symlinkHeaders[j].name, symlinkHeaders[j].linkname)
	})

	for _, h := range symlinkHeaders {
		evalPath := linknameFullPath(h.path, h.linkname)
		// Don't use constucted link if the link is absolute
		if filepath.IsAbs(h.linkname) {
			evalPath = h.linkname
		}

		// Check to see if the file that will be linked to is valid for symlinking
		_, err := filepath.EvalSymlinks(evalPath)
		if err != nil {
			return fmt.Errorf("failed to evaluate symlink %s: %w", h.path, err)
		}

		err = os.Symlink(h.linkname, h.path)
		if err != nil {
			return fmt.Errorf("failed to unzip symlink: %w", err)
		}
	}

	return nil
}

// This function checks to see that the given path is within the destination
// directory
func checkExtractPath(tarFilePath string, destination string) error {
	osPath := filepath.FromSlash(tarFilePath)
	destpath := filepath.Join(destination, osPath)
	if !strings.HasPrefix(destpath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("illegal file path %q: the file path does not occur within the destination directory", tarFilePath)
	}
	return nil
}

// Generates the full path for a symlink from the linkname and the symlink path
func linknameFullPath(path, linkname string) string {
	return filepath.Clean(filepath.Join(filepath.Dir(path), linkname))
}
