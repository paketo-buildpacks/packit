package servicebindings

import (
	"os"
)

// Entry represents the read-only content of a binding entry.
type Entry struct {
	path string
	file *os.File
}

// NewEntry returns a new Entry whose content is given by the file at the provided path.
func NewEntry(path string) *Entry {
	return &Entry{
		path: path,
	}
}

// ReadBytes reads the entire raw content of the entry. There is no need to call Close after calling ReadBytes.
func (e *Entry) ReadBytes() ([]byte, error) {
	return os.ReadFile(e.path)
}

// ReadString reads the entire content of the entry as a string. There is no need to call Close after calling
// ReadString.
func (e *Entry) ReadString() (string, error) {
	bytes, err := e.ReadBytes()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Read reads up to len(b) bytes from the entry. It returns the number of bytes read and any error encountered. At end
// of entry data, Read returns 0, io.EOF.
// Close must be called when all read operations are complete.
func (e *Entry) Read(b []byte) (int, error) {
	if e.file == nil {
		file, err := os.Open(e.path)
		if err != nil {
			return 0, err
		}
		e.file = file
	}
	return e.file.Read(b)
}

// Close closes the entry and resets it for reading. After calling Close, any subsequent calls to Read will read entry
// data from the beginning. Close may be called on a closed entry without error.
func (e *Entry) Close() error {
	if e.file == nil {
		return nil
	}
	defer func() {
		e.file = nil
	}()
	return e.file.Close()
}
