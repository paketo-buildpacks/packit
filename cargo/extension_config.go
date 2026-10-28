package cargo

import (
	"encoding/json"
	"io"
	"time"

	"github.com/BurntSushi/toml"
)

type ExtensionConfig struct {
	API       string                  `toml:"api"       json:"api,omitempty"`
	Extension ConfigExtension         `toml:"extension" json:"extension,omitempty"`
	Metadata  ConfigExtensionMetadata `toml:"metadata"  json:"metadata,omitempty"`
	Targets   []ConfigTarget          `toml:"targets"   json:"targets,omitempty"`
}

type ConfigExtensionMetadata struct {
	IncludeFiles    []string                               `toml:"include-files"              json:"include-files,omitempty"`
	PrePackage      string                                 `toml:"pre-package"                json:"pre-package,omitempty"`
	DefaultVersions map[string]string                      `toml:"default-versions"           json:"default-versions,omitempty"`
	Dependencies    []ConfigExtensionMetadataDependency    `toml:"dependencies"               json:"dependencies,omitempty"`
	Configurations  []ConfigExtensionMetadataConfiguration `toml:"configurations"             json:"configurations,omitempty"`
}

// ConfigExtensionMetadataDependency mirrors ConfigMetadataDependency so that
// extension.toml dependencies follow the same standard dependency metadata
// format as buildpack.toml dependencies.
type ConfigExtensionMetadataDependency struct {
	Arch     string `toml:"arch"             json:"arch,omitempty"`
	Checksum string `toml:"checksum"         json:"checksum,omitempty"`

	// CPE is the Common Platform Enumerator for the dependency.
	//
	// Deprecated: use CPEs instead.
	CPE string `toml:"cpe"              json:"cpe,omitempty"`

	// CPEs are the Common Platform Enumerators for the dependency.
	CPEs    []string       `toml:"cpes"             json:"cpes,omitempty"`
	Distros []ConfigDistro `toml:"distros"          json:"distros,omitempty"`

	// DeprecationDate is the date after which the dependency is considered
	// deprecated.
	//
	// Deprecated: use EOLDate instead.
	DeprecationDate *time.Time `toml:"deprecation_date" json:"deprecation_date,omitempty"`

	// EOLDate is the date after which the dependency is considered end of life.
	EOLDate *time.Time `toml:"eol-date"         json:"eol-date,omitempty"`

	ID string `toml:"id"               json:"id,omitempty"`

	// Licenses is the legacy representation of the dependency licenses, as a
	// list of license identifiers.
	//
	// Deprecated: use LicenseEntries instead.
	Licenses []interface{} `toml:"-" json:"-"`

	// LicenseEntries is the standard representation of the dependency licenses:
	// an array of tables, each with a type and an optional uri.
	//
	// Licenses and LicenseEntries share the single licenses key, so they are
	// (un)marshalled by the custom methods below rather than by struct tags.
	LicenseEntries []ConfigDependencyLicense `toml:"-" json:"-"`

	Name string `toml:"name"             json:"name,omitempty"`
	OS   string `toml:"os"               json:"os,omitempty"`

	// PURL is the package URL for the dependency.
	//
	// Deprecated: use PURLs instead.
	PURL string `toml:"purl"             json:"purl,omitempty"`

	// PURLs are the package URLs for the dependency.
	PURLs []string `toml:"purls"            json:"purls,omitempty"`

	SHA256         string   `toml:"sha256"           json:"sha256,omitempty"`
	Source         string   `toml:"source"           json:"source,omitempty"`
	SourceChecksum string   `toml:"source-checksum"  json:"source-checksum,omitempty"`
	Stacks         []string `toml:"stacks"           json:"stacks,omitempty"`

	// SourceSHA256 is the hex-encoded SHA256 checksum of the source-code
	// representation of the dependency. It is also read from the legacy
	// source_sha256 spelling, which used an underscore separator.
	//
	// Deprecated: use SourceChecksum instead.
	SourceSHA256    string `toml:"source-sha256"    json:"source-sha256,omitempty"`
	StripComponents int    `toml:"strip-components" json:"strip-components,omitempty"`
	URI             string `toml:"uri"              json:"uri,omitempty"`
	Version         string `toml:"version"          json:"version,omitempty"`
}

// MarshalJSON encodes the dependency, collapsing the two license
// representations onto the single licenses key.
func (d ConfigExtensionMetadataDependency) MarshalJSON() ([]byte, error) {
	type alias ConfigExtensionMetadataDependency

	return marshalDependency(alias(d), d.Licenses, d.LicenseEntries)
}

// UnmarshalJSON decodes the single licenses key into whichever of Licenses or
// LicenseEntries matches the representation used in the document.
func (d *ConfigExtensionMetadataDependency) UnmarshalJSON(data []byte) error {
	type alias ConfigExtensionMetadataDependency

	var decoded alias

	legacySourceSHA256, licenses, entries, err := unmarshalDependency(data, &decoded)
	if err != nil {
		return err
	}

	*d = ConfigExtensionMetadataDependency(decoded)
	d.Licenses = licenses
	d.LicenseEntries = entries

	// The source hash was previously written with an underscore separator.
	if d.SourceSHA256 == "" {
		d.SourceSHA256 = legacySourceSHA256
	}

	return nil
}

type ConfigExtensionMetadataConfiguration struct {
	Default     string `toml:"default"          json:"default,omitempty"`
	Launch      bool   `toml:"launch"           json:"launch,omitempty"`
	Description string `toml:"description"      json:"description,omitempty"`
	Build       bool   `toml:"build"            json:"build,omitempty"`
	Name        string `toml:"name"             json:"name,omitempty"`
}
type ConfigExtension struct {
	ID          string                   `toml:"id"                    json:"id,omitempty"`
	Name        string                   `toml:"name"                  json:"name,omitempty"`
	Version     string                   `toml:"version"               json:"version,omitempty"`
	Homepage    string                   `toml:"homepage,omitempty"    json:"homepage,omitempty"`
	Description string                   `toml:"description,omitempty" json:"description,omitempty"`
	Keywords    []string                 `toml:"keywords,omitempty"    json:"keywords,omitempty"`
	Licenses    []ConfigBuildpackLicense `toml:"licenses,omitempty"    json:"licenses,omitempty"`
	SBOMFormats []string                 `toml:"sbom-formats,omitempty"    json:"sbom-formats,omitempty"`
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

	c, err = convertExtensionStripComponents(extensionConfig.Metadata.Dependencies, c)
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

func (cd ConfigExtensionMetadataDependency) HasStack(stack string) bool {
	for _, s := range cd.Stacks {
		if s == stack {
			return true
		}
	}

	return false
}
