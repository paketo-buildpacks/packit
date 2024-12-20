package postal

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

type Checksum = cargo.Checksum

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

	// DeprecationDate is the data upon which this dependency is considered deprecated.
	DeprecationDate time.Time `toml:"deprecation_date"`

	// Checksum is a string that includes an algorithm and the hex-encoded hash
	// of the built dependency separated by a colon. Example
	// sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855.
	Checksum string `toml:"checksum"`

	// ID is the identifier used to specify the dependency.
	ID string `toml:"id"`

	// Licenses is a list of SPDX license identifiers of licenses in the dependency.
	Licenses []string `toml:"licenses"`

	// Name is the human-readable name of the dependency.
	Name string `toml:"name"`

	// PURL is the package URL for the dependency.
	PURL string `toml:"purl"`

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

	// ARCH is the architecture of this dependency
	Arch string `toml:"arch"`
}

func parseBuildpack(path, name string) ([]Dependency, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse buildpack.toml: %w", err)
	}

	var buildpack struct {
		Metadata struct {
			DefaultVersions map[string]string `toml:"default-versions"`
			Dependencies    []Dependency      `toml:"dependencies"`
		} `toml:"metadata"`
	}
	_, err = toml.NewDecoder(file).Decode(&buildpack)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse buildpack.toml: %w", err)
	}

	return buildpack.Metadata.Dependencies, buildpack.Metadata.DefaultVersions[name], nil
}

func stacksInclude(stacks []string, stack string) bool {
	for _, s := range stacks {
		if s == stack || s == "*" {
			return true
		}
	}
	return false
}
