package cargo

import (
	"encoding/json"
	"io"

	"github.com/BurntSushi/toml"
)

type ExtensionConfig struct {
	API       string                  `toml:"api"       json:"api,omitempty"`
	Extension ConfigExtension         `toml:"extension" json:"extension,omitempty"`
	Metadata  ConfigExtensionMetadata `toml:"metadata"  json:"metadata,omitempty"`
}

type ConfigExtensionMetadata struct {
	IncludeFiles    []string                            `toml:"include-files"              json:"include-files,omitempty"`
	PrePackage      string                              `toml:"pre-package"                json:"pre-package,omitempty"`
	DefaultVersions map[string]string                   `toml:"default-versions"           json:"default-versions,omitempty"`
	Dependencies    []ConfigExtensionMetadataDependency `toml:"dependencies"               json:"dependencies,omitempty"`
}

type ConfigExtensionMetadataDependency struct {
	Checksum       string        `toml:"checksum"         json:"checksum,omitempty"`
	ID             string        `toml:"id"               json:"id,omitempty"`
	Licenses       []interface{} `toml:"licenses"         json:"licenses,omitempty"`
	Name           string        `toml:"name"             json:"name,omitempty"`
	SHA256         string        `toml:"sha256"           json:"sha256,omitempty"`
	Source         string        `toml:"source"           json:"source,omitempty"`
	SourceChecksum string        `toml:"source-checksum"  json:"source-checksum,omitempty"`
	SourceSHA256   string        `toml:"source_sha256"    json:"source_sha256,omitempty"`
	Stacks         []string      `toml:"stacks"           json:"stacks,omitempty"`
	URI            string        `toml:"uri"              json:"uri,omitempty"`
	Version        string        `toml:"version"          json:"version,omitempty"`
}
type ConfigExtension struct {
	ID          string                   `toml:"id"                    json:"id,omitempty"`
	Name        string                   `toml:"name"                  json:"name,omitempty"`
	Version     string                   `toml:"version"               json:"version,omitempty"`
	Homepage    string                   `toml:"homepage,omitempty"    json:"homepage,omitempty"`
	Description string                   `toml:"description,omitempty" json:"description,omitempty"`
	Keywords    []string                 `toml:"keywords,omitempty"    json:"keywords,omitempty"`
	Licenses    []ConfigExtensionLicense `toml:"licenses,omitempty"    json:"licenses,omitempty"`
	SBOMFormats []string                 `toml:"sbom-formats,omitempty"    json:"sbom-formats,omitempty"`
}

type ConfigExtensionLicense struct {
	Type string `toml:"type" json:"type"`
	URI  string `toml:"uri"  json:"uri"`
}

func EncodeExtensionConfig(writer io.Writer, extensionConfig ExtensionConfig) error {
	content, err := json.Marshal(extensionConfig)
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

func DecodeExtensionConfig(reader io.Reader, extensionConfig *ExtensionConfig) error {
	var c map[string]interface{}
	_, err := toml.NewDecoder(reader).Decode(&c)
	if err != nil {
		return err
	}

	content, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return json.Unmarshal(content, extensionConfig)
}
