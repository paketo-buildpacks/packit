package cargo

import (
	"os"

	"github.com/pelletier/go-toml"
)

type Config struct {
	API       string          `toml:"api"`
	Buildpack ConfigBuildpack `toml:"buildpack"`
	Metadata  ConfigMetadata  `toml:"metadata"`
}

type ConfigBuildpack struct {
	ID      string `toml:"id"`
	Name    string `toml:"name"`
	Version string `toml:"version"`
}

type ConfigMetadata struct {
	IncludeFiles []string     `toml:"include_files"`
	PrePackage   string       `toml:"pre_package"`
	Dependencies []Dependency `toml:"dependencies"`
}

type Dependency struct {
	ID      string   `toml:"id"`
	Name    string   `toml:"name"`
	Sha256  string   `toml:"sha256"`
	Stacks  []string `toml:"stacks"`
	Uri     string   `toml:"uri"`
	Version string   `toml:"version"`
}

type BuildpackParser struct{}

func NewBuildpackParser() BuildpackParser {
	return BuildpackParser{}
}

func (p BuildpackParser) Parse(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = toml.NewDecoder(file).Decode(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
