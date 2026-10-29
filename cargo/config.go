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
	Targets   []ConfigTarget  `toml:"targets"   json:"targets,omitempty"`
}

type ConfigStack struct {
	ID     string   `toml:"id"     json:"id,omitempty"`
	Mixins []string `toml:"mixins" json:"mixins,omitempty"`
}

type ConfigTarget struct {
	OS   string `toml:"os"       json:"os,omitempty"`
	Arch string `toml:"arch"     json:"arch,omitempty"`
}

type ConfigDistro struct {
	Name    string `toml:"name"              json:"name"`
	Version string `toml:"version,omitempty" json:"version,omitempty"`
}

// ConfigDependencyLicense is a license applied to a dependency, in the standard
// dependency metadata format: an array of tables, each with a required type and
// an optional uri.
type ConfigDependencyLicense struct {
	Type string `toml:"type,omitempty" json:"type,omitempty"`
	URI  string `toml:"uri,omitempty"  json:"uri,omitempty"`
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
	Arch     string `toml:"arch"             json:"arch,omitempty"`
	Checksum string `toml:"checksum"         json:"checksum,omitempty"`

	// CPE is the Common Platform Enumerator for the dependency.
	//
	// Deprecated: use CPEs instead.
	CPE string `toml:"cpe"              json:"cpe,omitempty"`

	// CPEs are the Common Platform Enumerators for the dependency.
	CPEs    []string       `toml:"cpes"             json:"cpes,omitempty"`
	Distros []ConfigDistro `toml:"distros"          json:"distros,omitempty"`

	// PURL is the package URL for the dependency.
	//
	// Deprecated: use PURLs instead.
	PURL string `toml:"purl"             json:"purl,omitempty"`

	// PURLs are the package URLs for the dependency.
	PURLs []string `toml:"purls"            json:"purls,omitempty"`

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

	Name           string `toml:"name"             json:"name,omitempty"`
	OS             string `toml:"os"               json:"os,omitempty"`
	SHA256         string `toml:"sha256"           json:"sha256,omitempty"`
	Source         string `toml:"source"           json:"source,omitempty"`
	SourceChecksum string `toml:"source-checksum"  json:"source-checksum,omitempty"`

	// SourceSHA256 is the hex-encoded SHA256 checksum of the source-code
	// representation of the dependency. It is also read from the legacy
	// source_sha256 spelling, which used an underscore separator.
	//
	// Deprecated: use SourceChecksum instead.
	SourceSHA256    string   `toml:"source-sha256"    json:"source-sha256,omitempty"`
	Stacks          []string `toml:"stacks"           json:"stacks,omitempty"`
	StripComponents int      `toml:"strip-components" json:"strip-components,omitempty"`
	URI             string   `toml:"uri"              json:"uri,omitempty"`
	Version         string   `toml:"version"          json:"version,omitempty"`
}

// marshalDependency encodes a dependency struct, collapsing the two license
// representations onto the single licenses key.
//
// The standard table form is emitted when entries is set. Otherwise the legacy
// list of identifiers is emitted, so buildpack.toml and extension.toml files
// that use the older format round-trip unchanged.
func marshalDependency(value any, legacy []interface{}, entries []ConfigDependencyLicense) ([]byte, error) {
	content, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var encoded map[string]interface{}
	if err := json.Unmarshal(content, &encoded); err != nil {
		return nil, err
	}

	switch {
	case len(entries) > 0:
		encoded["licenses"] = entries
	case len(legacy) > 0:
		encoded["licenses"] = legacy
	}

	return json.Marshal(encoded)
}

// unmarshalDependency decodes a dependency struct and splits the single
// licenses key into whichever representation the document uses. It also returns
// the value of the legacy source_sha256 key, which used an underscore separator.
func unmarshalDependency(data []byte, value any) (string, []interface{}, []ConfigDependencyLicense, error) {
	if err := json.Unmarshal(data, value); err != nil {
		return "", nil, nil, err
	}

	var raw struct {
		Licenses           []json.RawMessage `json:"licenses"`
		LegacySourceSHA256 string            `json:"source_sha256"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", nil, nil, err
	}

	var licenses []interface{}
	var entries []ConfigDependencyLicense

	for _, item := range raw.Licenses {
		var identifier string
		if err := json.Unmarshal(item, &identifier); err == nil {
			licenses = append(licenses, identifier)
			continue
		}

		var entry ConfigDependencyLicense
		if err := json.Unmarshal(item, &entry); err != nil {
			return "", nil, nil, err
		}

		entries = append(entries, entry)
	}

	return raw.LegacySourceSHA256, licenses, entries, nil
}

// MarshalJSON encodes the dependency, collapsing the two license
// representations onto the single licenses key.
//
// The standard table form is emitted when LicenseEntries is set. Otherwise the
// legacy list of identifiers is emitted, so buildpack.toml files that use the
// older format round-trip unchanged.
func (d ConfigMetadataDependency) MarshalJSON() ([]byte, error) {
	type alias ConfigMetadataDependency

	return marshalDependency(alias(d), d.Licenses, d.LicenseEntries)
}

// UnmarshalJSON decodes the single licenses key into whichever of Licenses or
// LicenseEntries matches the representation used in the document.
func (d *ConfigMetadataDependency) UnmarshalJSON(data []byte) error {
	type alias ConfigMetadataDependency

	var decoded alias

	legacySourceSHA256, licenses, entries, err := unmarshalDependency(data, &decoded)
	if err != nil {
		return err
	}

	*d = ConfigMetadataDependency(decoded)
	d.Licenses = licenses
	d.LicenseEntries = entries

	// The source hash was previously written with an underscore separator.
	if d.SourceSHA256 == "" {
		d.SourceSHA256 = legacySourceSHA256
	}

	return nil
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

	c, err = convertStripComponents(config.Metadata.Dependencies, c)
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
func convertPatches(cons []ConfigMetadataDependencyConstraint, c map[string]interface{}) (map[string]interface{}, error) {
	if len(cons) > 0 {
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

// Accomplishes the same this as the convertPatches function but for strip components in the dependencies list.
func convertStripComponents(deps []ConfigMetadataDependency, c map[string]interface{}) (map[string]interface{}, error) {
	if len(deps) == 0 {
		return c, nil
	}

	return convertDependencyStripComponents(c)
}

func convertExtensionStripComponents(deps []ConfigExtensionMetadataDependency, c map[string]interface{}) (map[string]interface{}, error) {
	if len(deps) == 0 {
		return c, nil
	}

	return convertDependencyStripComponents(c)
}

func convertDependencyStripComponents(c map[string]interface{}) (map[string]interface{}, error) {
	metadata, ok := c["metadata"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failure to assert type: unexpected data in metadata")
	}

	dependencies, ok := metadata["dependencies"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("failure to assert type: unexpected data in constraints")
	}

	for _, dependency := range dependencies {
		stripComponents, ok := dependency.(map[string]interface{})["strip-components"]
		if !ok {
			continue
		}

		floatStripComponents, ok := stripComponents.(float64)
		if !ok {
			return nil, fmt.Errorf("failure to assert type: unexpected data")
		}
		dependency.(map[string]interface{})["strip-components"] = int(floatStripComponents)
	}

	return c, nil
}
