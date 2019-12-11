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
	IncludeFiles []string `toml:"include_files"`
	PrePackage   string   `toml:"pre_package"`
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
