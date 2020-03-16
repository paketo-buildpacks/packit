package postal

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

// Dependency is a representation of a buildpack dependency.
type Dependency struct {
	// DeprecationDate is the data upon which this dependency is considered deprecated.
	DeprecationDate time.Time `toml:"deprecation_date"`

	// ID is the identifier used to specify the dependency.
	ID string `toml:"id"`

	// Version is the specific version of the dependency.
	Version string `toml:"version"`

	// Name is the human-readable name of the dependency.
	Name string `toml:"name"`

	// URI is the uri location of the built dependency.
	URI string `toml:"uri"`

	// SHA256 is the hex-encoded SHA256 checksum of the built dependency.
	SHA256 string `toml:"sha256"`

	// Source is the uri location of the source-code representation of the dependency.
	Source string `toml:"source"`

	// SourceSHA256 is the hex-encoded SHA256 checksum of the source-code representation of the dependency.
	SourceSHA256 string `toml:"source_sha256"`

	// Stacks is a list of stacks for which the dependency is built.
	Stacks []string `toml:"stacks"`
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
	_, err = toml.DecodeReader(file, &buildpack)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse buildpack.toml: %w", err)
	}

	return buildpack.Metadata.Dependencies, buildpack.Metadata.DefaultVersions[name], nil
}

func stacksInclude(stacks []string, stack string) bool {
	for _, s := range stacks {
		if s == stack {
			return true
		}
	}
	return false
}
