package packit

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type Environment map[string]string

func NewEnvironment() Environment {
	return Environment{}
}

func NewEnvironmentFromPath(path string) (Environment, error) {
	envFiles, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return Environment{}, fmt.Errorf("failed to match env directory files: %s", err)
	}

	environment := Environment{}
	for _, file := range envFiles {
		switch filepath.Ext(file) {
		case ".delim", ".prepend", ".append", ".default", ".override":
			contents, err := ioutil.ReadFile(file)
			if err != nil {
				return Environment{}, fmt.Errorf("failed to load environment variable: %s", err)
			}

			environment[filepath.Base(file)] = string(contents)
		}
	}

	return environment, nil
}

func (e Environment) Override(name, value string) {
	e[name+".override"] = value
}

func (e Environment) Prepend(name, value, delim string) {
	e[name+".prepend"] = value

	delete(e, name+".delim")
	if delim != "" {
		e[name+".delim"] = delim
	}
}

func (e Environment) Append(name, value, delim string) {
	e[name+".append"] = value

	delete(e, name+".delim")
	if delim != "" {
		e[name+".delim"] = delim
	}
}

func (e Environment) Default(name, value string) {
	e[name+".default"] = value
}
