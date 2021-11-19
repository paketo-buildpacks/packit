package internal

import (
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
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}

	return nil
}
