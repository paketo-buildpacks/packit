package vacation

import (
	"io"
	"os"
	"path/filepath"
)

// An Executable writes an executable files from an input stream to the bin/ directory.
type Executable struct {
	reader     io.Reader
	name string
}

func NewExecutable(inputReader io.Reader, name string) Executable {
	return Executable{reader: inputReader, name: name}
}

func (e Executable) Decompress(destination string) error {
	err := os.MkdirAll(filepath.Join(destination, "bin"), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(destination, "bin", e.name))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, e.reader)
	if err != nil {
		return err
	}

	err = os.Chmod(filepath.Join(destination, "bin", e.name), 0755)
	if err != nil {
		return err
	}

	return nil
}