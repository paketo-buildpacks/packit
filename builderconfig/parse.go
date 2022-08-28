package builderconfig

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type metadata struct {
	Disable bool   `toml:"disable-builder-config"`
	Path    string `toml:"builder-config-path"`
}

func ParseBuildpackAndResolve(path string) error {
	buildpackMetadata := struct {
		Metadata metadata `toml:"metadata"`
	}{}
	if _, err := toml.DecodeFile(path, &buildpackMetadata); err != nil {
		return fmt.Errorf("unable to parse buildpack.toml: %w", err)
	}
	if buildpackMetadata.Metadata.Disable {
		return nil
	}
	return New(WithPath(buildpackMetadata.Metadata.Path)).Resolve()
}
