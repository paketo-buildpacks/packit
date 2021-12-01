package vacation

import (
	"compress/bzip2"
	"io"
)

// A Bzip2Archive decompresses bzip2 files from an input stream.
type Bzip2Archive struct {
	reader     io.Reader
	components int
}

// NewBzip2Archive returns a new Bzip2Archive that reads from inputReader.
func NewBzip2Archive(inputReader io.Reader) Bzip2Archive {
	return Bzip2Archive{reader: inputReader}
}

// Decompress reads from Bzip2Archive and writes files into the destination
// specified.
func (tbz Bzip2Archive) Decompress(destination string) error {
	return NewArchive(bzip2.NewReader(tbz.reader)).StripComponents(tbz.components).Decompress(destination)
}

// StripComponents behaves like the --strip-components flag on tar command
// removing the first n levels from the final decompression destination.
func (tbz Bzip2Archive) StripComponents(components int) Bzip2Archive {
	tbz.components = components
	return tbz
}
