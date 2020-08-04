package packit

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// Environment provides a key-value store for declaring environment variables.
type Environment map[string]string

// Append adds a key-value pair to the environment as an appended value
// according to the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#append.
func (e Environment) Append(name, value, delim string) {
	e[name+".append"] = value

	delete(e, name+".delim")
	if delim != "" {
		e[name+".delim"] = delim
	}
}

// Default adds a key-value pair to the environment as a default value
// according to the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#default.
func (e Environment) Default(name, value string) {
	e[name+".default"] = value
}

// Override adds a key-value pair to the environment as an overridden value
// according to the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#override.
func (e Environment) Override(name, value string) {
	e[name+".override"] = value
}

// Prepend adds a key-value pair to the environment as a prepended value
// according to the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#prepend.
func (e Environment) Prepend(name, value, delim string) {
	e[name+".prepend"] = value

	delete(e, name+".delim")
	if delim != "" {
		e[name+".delim"] = delim
	}
}

func newEnvironmentFromPath(path string) (Environment, error) {
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
