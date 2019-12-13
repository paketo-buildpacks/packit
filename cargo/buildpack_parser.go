package cargo

import "os"

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
	err = DecodeConfig(file, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
