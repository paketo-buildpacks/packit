package internal

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type EnvironmentWriter struct{}

func NewEnvironmentWriter() EnvironmentWriter {
	return EnvironmentWriter{}
}

func (w EnvironmentWriter) Write(dir string, env map[string]string) error {
	if len(env) == 0 {
		return nil
	}

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	for key, value := range env {
		err := ioutil.WriteFile(filepath.Join(dir, key), []byte(value), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
