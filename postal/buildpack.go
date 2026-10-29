package postal

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

type Checksum = cargo.Checksum
type Distro = cargo.ConfigDistro

// License is a license applied to a dependency, in the standard dependency
// metadata format: a table with a required type and an optional uri.
type License struct {
	Type string `toml:"type"`
	URI  string `toml:"uri"`
}

// Dependency is a representation of a buildpack dependency.
type Dependency struct {
	// CPE is the Common Platform Enumerator for the dependency. Used in legacy
	// image label SBOM (GenerateBillOfMaterials).
	//
	// Deprecated: use CPEs instead.
	CPE string `toml:"cpe"`

	// CPEs are the Common Platform Enumerators for the dependency. Used in Syft
	// and SPDX JSON SBOMs. If unset, falls back to CPE.
	CPEs []string `toml:"cpes"`

	// DeprecationDate is the date after which the dependency is considered
	// deprecated. It is populated from the same value as EOLDate, and is kept so
	// that existing consumers keep working. Prefer EOLDate in new code.
	DeprecationDate time.Time `toml:"deprecation_date"`

	// EOLDate is the date after which the dependency is considered end of life.
	// It is read from the standard eol-date key, and from the legacy
	// deprecation_date key for backwards compatibility.
	EOLDate time.Time `toml:"eol-date"`

	// Checksum is a string that includes an algorithm and the hex-encoded hash
	// of the built dependency separated by a colon. Example
	// sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855.
	Checksum string `toml:"checksum"`

	// ID is the identifier used to specify the dependency.
	ID string `toml:"id"`

	// Licenses is the legacy representation of the dependency licenses, as a
	// list of license identifiers.
	//
	// Deprecated: use LicenseEntries instead.
	Licenses []string `toml:"-"`

	// LicenseEntries is the standard representation of the dependency licenses:
	// a list of tables, each with a type and an optional uri.
	//
	// Licenses and LicenseEntries share the single licenses key, so they are
	// populated by parseBuildpack rather than by struct tags.
	LicenseEntries []License `toml:"-"`

	// Name is the human-readable name of the dependency.
	Name string `toml:"name"`

	// PURL is the package URL for the dependency.
	//
	// Deprecated: use PURLs instead.
	PURL string `toml:"purl"`

	// PURLs are the package URLs for the dependency.
	PURLs []string `toml:"purls"`

	// SHA256 is the hex-encoded SHA256 checksum of the built dependency.
	//
	// Deprecated: use Checksum instead.
	SHA256 string `toml:"sha256"`

	// Source is the uri location of the source-code representation of the dependency.
	Source string `toml:"source"`

	// SourceChecksum is a string that includes an algorithm and the hex-encoded
	// hash of the source representation of the dependency separated by a colon.
	// Example sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855.
	SourceChecksum string `toml:"source-checksum"`

	// SourceSHA256 is the hex-encoded SHA256 checksum of the source-code
	// representation of the dependency.
	//
	// Deprecated: use SourceChecksum instead.
	SourceSHA256 string `toml:"source_sha256"`

	// Stacks is a list of stacks for which the dependency is built.
	Stacks []string `toml:"stacks"`

	// URI is the uri location of the built dependency.
	URI string `toml:"uri"`

	// Version is the specific version of the dependency.
	Version string `toml:"version"`

	// StripComponents behaves like the --strip-components flag on tar command
	// removing the first n levels from the final decompression destination.
	StripComponents int `toml:"strip-components"`

	// Platform information of the dependency
	OS   string `toml:"os"`
	Arch string `toml:"arch"`

	// Optional list of OS distribution supported by the dependency.
	Distros []Distro `toml:"distros"`
}

// dependencyMetadata mirrors Dependency and additionally captures keys that
// share a name with a struct field but use a different spelling, or that can
// take more than one form. It is decoded separately from Dependency.
type dependencyMetadata struct {
	Dependency

	RawLicenses []any `toml:"licenses"`

	// SourceSHA256Dashed is the hyphenated spelling of source_sha256, which is
	// what buildpack.toml files in the wild use.
	SourceSHA256Dashed string `toml:"source-sha256"`
}

func parseBuildpack(path, name string) ([]Dependency, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse buildpack.toml: %w", err)
	}

	var buildpack struct {
		Metadata struct {
			DefaultVersions map[string]string    `toml:"default-versions"`
			Dependencies    []dependencyMetadata `toml:"dependencies"`
		} `toml:"metadata"`
	}
	_, err = toml.NewDecoder(file).Decode(&buildpack)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse buildpack.toml: %w", err)
	}

	dependencies := make([]Dependency, 0, len(buildpack.Metadata.Dependencies))
	for _, metadata := range buildpack.Metadata.Dependencies {
		dependency := metadata.Dependency

		if err := dependency.setLicenses(metadata.RawLicenses); err != nil {
			return nil, "", fmt.Errorf("failed to parse buildpack.toml: %w", err)
		}

		if dependency.SourceSHA256 == "" {
			dependency.SourceSHA256 = metadata.SourceSHA256Dashed
		}

		dependency.normalize()
		dependencies = append(dependencies, dependency)
	}

	return dependencies, buildpack.Metadata.DefaultVersions[name], nil
}

// setLicenses splits the raw licenses value into the representation that
// matches the document. A bare string is the legacy form, a table is the
// standard form.
func (d *Dependency) setLicenses(values []any) error {
	for _, value := range values {
		switch license := value.(type) {
		case string:
			d.Licenses = append(d.Licenses, license)

		case map[string]any:
			entry := License{}
			entry.Type, _ = license["type"].(string)
			entry.URI, _ = license["uri"].(string)

			d.LicenseEntries = append(d.LicenseEntries, entry)

		default:
			return fmt.Errorf("unexpected license format: %T", value)
		}
	}

	return nil
}

// normalize keeps the legacy and standard fields for the same value in sync so
// that consumers can rely on either one being populated, regardless of which
// key was used in buildpack.toml.
func (d *Dependency) normalize() {
	if d.EOLDate.IsZero() {
		d.EOLDate = d.DeprecationDate
	}

	if d.DeprecationDate.IsZero() {
		d.DeprecationDate = d.EOLDate
	}
}

func stacksInclude(stacks []string, stack string) bool {
	for _, s := range stacks {
		if s == stack || s == "*" {
			return true
		}
	}
	return false
}

func supportsPlatform(targetOs, targetArch string, dependency Dependency) bool {

	// Avoid strict checking in case of dependency does not specify OS/Arch
	if dependency.OS == "" && dependency.Arch == "" {
		return true
	}

	if targetOs != dependency.OS || targetArch != dependency.Arch {
		return false
	}

	return true
}
