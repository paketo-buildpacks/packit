package cargo

import "os"

type ExtensionParser struct{}

func NewExtensionParser() ExtensionParser {
	return ExtensionParser{}
}

func (p ExtensionParser) Parse(path string) (Config, error) {
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
