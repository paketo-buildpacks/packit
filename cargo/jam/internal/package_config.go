package internal

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/pelletier/go-toml"
)

type PackageConfig struct {
	Buildpack    interface{}               `toml:"buildpack"`
	Dependencies []PackageConfigDependency `toml:"dependencies"`
}

type PackageConfigDependency struct {
	URI string `toml:"uri"`
}

// Note: this is to support that buildpackages can refer to this field as `image` or `uri`.
func (d *PackageConfigDependency) UnmarshalTOML(v interface{}) error {
	if m, ok := v.(map[string]interface{}); ok {
		if image, ok := m["image"].(string); ok {
			d.URI = image
		}

		if uri, ok := m["uri"].(string); ok {
			d.URI = uri
		}
	}

	if d.URI != "" {
		uri, err := url.Parse(d.URI)
		if err != nil {
			return err
		}

		uri.Scheme = ""

		d.URI = strings.TrimPrefix(uri.String(), "//")
	}

	return nil
}

func ParsePackageConfig(path string) (PackageConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return PackageConfig{}, fmt.Errorf("failed to open package config file: %w", err)
	}
	defer file.Close()

	var config PackageConfig
	err = toml.NewDecoder(file).Decode(&config)
	if err != nil {
		return PackageConfig{}, fmt.Errorf("failed to parse package config: %w", err)
	}

	return config, nil
}

func OverwritePackageConfig(path string, config PackageConfig) error {
	for i, dependency := range config.Dependencies {
		if !strings.HasPrefix(dependency.URI, "docker://") {
			config.Dependencies[i].URI = fmt.Sprintf("docker://%s", dependency.URI)
		}
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open package config file: %w", err)
	}

	err = toml.NewEncoder(file).Encode(config)
	if err != nil {
		return fmt.Errorf("failed to write package config: %w", err)
	}

	return nil
}
