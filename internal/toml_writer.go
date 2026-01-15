package internal

import (
	"errors"
	"os"

	"github.com/pelletier/go-toml"
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

	var errs error

	defer func() {
		errs = errors.Join(errs, file.Close())
	}()

	return errors.Join(errs, toml.NewEncoder(file).Encode(value))
}
