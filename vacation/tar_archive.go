package vacation

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// A TarArchive decompresses tar files from an input stream.
type TarArchive struct {
	reader     io.Reader
	components int
}

// NewTarArchive returns a new TarArchive that reads from inputReader.
func NewTarArchive(inputReader io.Reader) TarArchive {
	return TarArchive{reader: inputReader}
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

		// Clean the name in the header to prevent './filename' being stripped to
		// 'filename' also to skip if the destination it the destination directory
		// itself i.e. './'
		var name string
		if name = filepath.Clean(hdr.Name); name == "." {
			continue
		}

		err = checkExtractPath(name, destination)
		if err != nil {
			return err
		}

		fileNames := strings.Split(name, "/")

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
				linkname: hdr.Linkname,
				path:     path,
			})
		}
	}

	// Create a map of all of the symlink names and where they are pointing to to
	// act as a quasi-graph
	symlinkMap := map[string]string{}
	for _, h := range symlinkHeaders {
		symlinkMap[filepath.Clean(h.path)] = h.linkname
	}

	// Iterate over the symlink map for every link that is found this ensures
	// that all symlinks that can be created will be created and any that are
	// left over are cyclically dependent
	maxIterations := len(symlinkMap)
	for i := 0; i < maxIterations; i++ {
		for path, linkname := range symlinkMap {
			// Check to see if the linkname lies on the path of another symlink in
			// the table or if it is another symlink in the table
			//
			// Example:
			// path = dir/file
			// a-symlink -> dir
			// b-symlink -> a-symlink
			// c-symlink -> a-symlink/file
			//
			// If there is a match either of the symlink or it is on the path then
			// skip the creation of this symlink for now
			shouldSkipLink := func() bool {
				sln := strings.Split(linkname, "/")
				for j := 0; j < len(sln); j++ {
					if _, ok := symlinkMap[linknameFullPath(path, filepath.Join(sln[:j+1]...))]; ok {
						return true
					}
				}
				return false
			}

			if shouldSkipLink() {
				continue
			}

			// If the linkname is not an existing link in the symlink table then we
			// can attempt the make the link

			// Check to see if the file that will be linked to is valid for symlinking
			_, err := filepath.EvalSymlinks(linknameFullPath(path, linkname))
			if err != nil {
				return fmt.Errorf("failed to evaluate symlink %s: %w", path, err)
			}

			// Create the symlink
			err = os.Symlink(linkname, path)
			if err != nil {
				return fmt.Errorf("failed to extract symlink: %s", err)
			}

			// Remove the created symlink from the symlink table so that its
			// dependent symlinks can be created in the next iteration
			delete(symlinkMap, path)
		}
	}

	// Check to see if there are any symlinks left in the map which would
	// indicate a cyclical dependency
	if len(symlinkMap) > 0 {
		return fmt.Errorf("failed: max iterations reached: this symlink graph contains a cycle")
	}

	return nil
}

// StripComponents behaves like the --strip-components flag on tar command
// removing the first n levels from the final decompression destination.
func (ta TarArchive) StripComponents(components int) TarArchive {
	ta.components = components
	return ta
}
