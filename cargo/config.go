package cargo

import (
	"encoding/json"
	"io"

	"github.com/BurntSushi/toml"
)

type Config struct {
	API       string          `toml:"api" json:"api"`
	Buildpack ConfigBuildpack `toml:"buildpack" json:"buildpack"`
	Metadata  ConfigMetadata  `toml:"metadata" json:"metadata"`
	Stacks    []ConfigStack   `toml:"stacks" json:"stacks"`
}

type ConfigStack struct {
	ID string `toml:"id" json:"id"`
}

type ConfigBuildpack struct {
	ID      string `toml:"id" json:"id"`
	Name    string `toml:"name" json:"name"`
	Version string `toml:"version" json:"version"`
}

type ConfigMetadata struct {
	IncludeFiles []string                   `toml:"include_files" json:"include_files"`
	PrePackage   string                     `toml:"pre_package" json:"pre_package"`
	Dependencies []ConfigMetadataDependency `toml:"dependencies" json:"dependencies"`
	Unstructured map[string]interface{}     `toml:"-" json:"-"`
}

type ConfigMetadataDependency struct {
	ID      string   `toml:"id" json:"id"`
	Name    string   `toml:"name" json:"name"`
	SHA256  string   `toml:"sha256" json:"sha256"`
	Stacks  []string `toml:"stacks" json:"stacks"`
	URI     string   `toml:"uri" json:"uri"`
	Version string   `toml:"version" json:"version"`
}

func EncodeConfig(writer io.Writer, config Config) error {
	content, err := json.Marshal(config)
	if err != nil {
		return err
	}

	c := map[string]interface{}{}
	err = json.Unmarshal(content, &c)
	if err != nil {
		return err
	}

	return toml.NewEncoder(writer).Encode(c)
}

func DecodeConfig(reader io.Reader, config *Config) error {
	var c map[string]interface{}
	_, err := toml.DecodeReader(reader, &c)
	if err != nil {
		return err
	}

	content, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return json.Unmarshal(content, config)
}

func (m ConfigMetadata) MarshalJSON() ([]byte, error) {
	metadata := map[string]interface{}{}

	for key, value := range m.Unstructured {
		metadata[key] = value
	}

	metadata["include_files"] = m.IncludeFiles
	metadata["pre_package"] = m.PrePackage
	metadata["dependencies"] = m.Dependencies

	return json.Marshal(metadata)
}

func (m *ConfigMetadata) UnmarshalJSON(data []byte) error {
	var metadata map[string]json.RawMessage
	err := json.Unmarshal(data, &metadata)
	if err != nil {
		return err
	}

	if includeFiles, ok := metadata["include_files"]; ok {
		err = json.Unmarshal(includeFiles, &m.IncludeFiles)
		if err != nil {
			return err
		}
		delete(metadata, "include_files")
	}

	if prePackage, ok := metadata["pre_package"]; ok {
		err = json.Unmarshal(prePackage, &m.PrePackage)
		if err != nil {
			return err
		}
		delete(metadata, "pre_package")
	}

	if dependencies, ok := metadata["dependencies"]; ok {
		err = json.Unmarshal(dependencies, &m.Dependencies)
		if err != nil {
			return err
		}
		delete(metadata, "dependencies")
	}

	if len(metadata) > 0 {
		m.Unstructured = map[string]interface{}{}
		for key, value := range metadata {
			m.Unstructured[key] = value
		}
	}

	return nil
}

func (cd ConfigMetadataDependency) HasStack(stack string) bool {
	for _, s := range cd.Stacks {
		if s == stack {
			return true
		}
	}

	return false
}
