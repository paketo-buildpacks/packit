package postal

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

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

type Dependency struct {
	ID           string `toml:"id"`
	Name         string `toml:"name"`
	SHA256       string `toml:"sha256"`
	Source       string `toml:"source"`
	SourceSHA256 string `toml:"source_sha256"`
	Stacks       Stacks `toml:"stacks"`
	URI          string `toml:"uri"`
	Version      string `toml:"version"`
}

type Stacks []string

func (stacks Stacks) Include(stack string) bool {
	for _, s := range stacks {
		if s == stack {
			return true
		}
	}
	return false
}
