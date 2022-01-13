package cargo

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	API       string          `toml:"api"       json:"api,omitempty"`
	Buildpack ConfigBuildpack `toml:"buildpack" json:"buildpack,omitempty"`
	Metadata  ConfigMetadata  `toml:"metadata"  json:"metadata,omitempty"`
	Stacks    []ConfigStack   `toml:"stacks"    json:"stacks,omitempty"`
	Order     []ConfigOrder   `toml:"order"     json:"order,omitempty"`
}

type ConfigStack struct {
	ID     string   `toml:"id"     json:"id,omitempty"`
	Mixins []string `toml:"mixins" json:"mixins,omitempty"`
}

type ConfigBuildpack struct {
	ID          string                   `toml:"id"                    json:"id,omitempty"`
	Name        string                   `toml:"name"                  json:"name,omitempty"`
	Version     string                   `toml:"version"               json:"version,omitempty"`
	Homepage    string                   `toml:"homepage,omitempty"    json:"homepage,omitempty"`
	ClearEnv    bool                     `toml:"clear-env,omitempty"   json:"clear-env,omitempty"`
	Description string                   `toml:"description,omitempty" json:"description,omitempty"`
	Keywords    []string                 `toml:"keywords,omitempty"    json:"keywords,omitempty"`
	Licenses    []ConfigBuildpackLicense `toml:"licenses,omitempty"    json:"licenses,omitempty"`
	SBOMFormats []string                 `toml:"sbom-formats,omitempty"    json:"sbom-formats,omitempty"`
}

type ConfigBuildpackLicense struct {
	Type string `toml:"type" json:"type"`
	URI  string `toml:"uri"  json:"uri"`
}

type ConfigMetadata struct {
	IncludeFiles          []string                             `toml:"include-files"              json:"include-files,omitempty"`
	PrePackage            string                               `toml:"pre-package"                json:"pre-package,omitempty"`
	DefaultVersions       map[string]string                    `toml:"default-versions"           json:"default-versions,omitempty"`
	Dependencies          []ConfigMetadataDependency           `toml:"dependencies"               json:"dependencies,omitempty"`
	DependencyConstraints []ConfigMetadataDependencyConstraint `toml:"dependency-constraints"     json:"dependency-constraints,omitempty"`
	Unstructured          map[string]interface{}               `toml:"-"                          json:"-"`
}

type ConfigMetadataDependency struct {
	CPE             string        `toml:"cpe"              json:"cpe,omitempty"`
	PURL            string        `toml:"purl"              json:"purl,omitempty"`
	DeprecationDate *time.Time    `toml:"deprecation_date" json:"deprecation_date,omitempty"`
	ID              string        `toml:"id"               json:"id,omitempty"`
	Licenses        []interface{} `toml:"licenses"         json:"licenses,omitempty"`
	Name            string        `toml:"name"             json:"name,omitempty"`
	SHA256          string        `toml:"sha256"           json:"sha256,omitempty"`
	Source          string        `toml:"source"           json:"source,omitempty"`
	SourceSHA256    string        `toml:"source_sha256"    json:"source_sha256,omitempty"`
	Stacks          []string      `toml:"stacks"           json:"stacks,omitempty"`
	URI             string        `toml:"uri"              json:"uri,omitempty"`
	Version         string        `toml:"version"          json:"version,omitempty"`
}

type ConfigMetadataDependencyConstraint struct {
	Constraint string `toml:"constraint"       json:"constraint,omitempty"`
	ID         string `toml:"id"               json:"id,omitempty"`
	Patches    int    `toml:"patches"          json:"patches,omitempty"`
}

type ConfigOrder struct {
	Group []ConfigOrderGroup `toml:"group" json:"group,omitempty"`
}

type ConfigOrderGroup struct {
	ID       string `toml:"id"       json:"id,omitempty"`
	Version  string `toml:"version"  json:"version,omitempty"`
	Optional bool   `toml:"optional,omitempty" json:"optional,omitempty"`
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

	c, err = convertPatches(config.Metadata.DependencyConstraints, c)
	if err != nil {
		return err
	}

	return toml.NewEncoder(writer).Encode(c)
}

func DecodeConfig(reader io.Reader, config *Config) error {
	var c map[string]interface{}
	_, err := toml.NewDecoder(reader).Decode(&c)
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

	if len(m.IncludeFiles) > 0 {
		metadata["include-files"] = m.IncludeFiles
	}

	if len(m.PrePackage) > 0 {
		metadata["pre-package"] = m.PrePackage
	}

	if len(m.Dependencies) > 0 {
		metadata["dependencies"] = m.Dependencies
	}

	if len(m.DependencyConstraints) > 0 {
		metadata["dependency-constraints"] = m.DependencyConstraints
	}

	if len(m.DefaultVersions) > 0 {
		metadata["default-versions"] = m.DefaultVersions
	}

	return json.Marshal(metadata)
}

func (m *ConfigMetadata) UnmarshalJSON(data []byte) error {
	var metadata map[string]json.RawMessage
	err := json.Unmarshal(data, &metadata)
	if err != nil {
		return err
	}

	if includeFiles, ok := metadata["include-files"]; ok {
		err = json.Unmarshal(includeFiles, &m.IncludeFiles)
		if err != nil {
			return err
		}
		delete(metadata, "include-files")
	}

	if prePackage, ok := metadata["pre-package"]; ok {
		err = json.Unmarshal(prePackage, &m.PrePackage)
		if err != nil {
			return err
		}
		delete(metadata, "pre-package")
	}

	if dependencies, ok := metadata["dependencies"]; ok {
		err = json.Unmarshal(dependencies, &m.Dependencies)
		if err != nil {
			return err
		}
		delete(metadata, "dependencies")
	}

	if dependencyConstraints, ok := metadata["dependency-constraints"]; ok {
		err = json.Unmarshal(dependencyConstraints, &m.DependencyConstraints)
		if err != nil {
			return err
		}
		delete(metadata, "dependency-constraints")
	}

	if defaultVersions, ok := metadata["default-versions"]; ok {
		err = json.Unmarshal(defaultVersions, &m.DefaultVersions)
		if err != nil {
			return err
		}
		delete(metadata, "default-versions")
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

// Unmarshal stores json numbers in float64 types, adding an unnecessary decimal point to the patch in the final toml.
// convertPatches converts this float64 into an int and returns a new map that contains an integer value for patches
func convertPatches(constraints []ConfigMetadataDependencyConstraint, c map[string]interface{}) (map[string]interface{}, error) {
	if len(constraints) > 0 {
		metadata, ok := c["metadata"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failure to assert type: unexpected data in metadata")
		}

		constraints, ok := metadata["dependency-constraints"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("failure to assert type: unexpected data in constraints")
		}

		for _, dependencyConstraint := range constraints {
			patches, ok := dependencyConstraint.(map[string]interface{})["patches"]
			if !ok {
				return nil, fmt.Errorf("failure to assert type: unexpected data in constraint patches")
			}

			floatPatches, ok := patches.(float64)
			if !ok {
				return nil, fmt.Errorf("failure to assert type: unexpected data")
			}
			dependencyConstraint.(map[string]interface{})["patches"] = int(floatPatches)
		}
	}
	return c, nil
}
