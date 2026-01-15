package internal

import (
	"errors"
	"io"
	"os"
)

type FileWriter struct{}

func NewFileWriter() FileWriter {
	return FileWriter{}
}

func (fw FileWriter) Write(path string, reader io.Reader) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	var errs error

	defer func() {
		errs = errors.Join(errs, file.Close())
	}()

	_, err = io.Copy(file, reader)
	return errors.Join(errs, err)
}
