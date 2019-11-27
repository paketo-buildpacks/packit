package internal

import (
	"os"

	"github.com/BurntSushi/toml"
)

type TOMLWriter struct{}

func NewTOMLWriter() TOMLWriter {
	return TOMLWriter{}
}

func (tw TOMLWriter) Write(path string, value interface{}) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return toml.NewEncoder(file).Encode(value)
}
