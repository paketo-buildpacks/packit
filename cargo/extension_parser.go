package cargo

import "os"

type ExtensionParser struct{}

func NewExtensionParser() ExtensionParser {
	return ExtensionParser{}
}

func (p ExtensionParser) Parse(path string) (ExtensionConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return ExtensionConfig{}, err
	}

	var config ExtensionConfig
	err = DecodeExtensionConfig(file, &config)
	if err != nil {
		return ExtensionConfig{}, err
	}

	return config, nil
}
