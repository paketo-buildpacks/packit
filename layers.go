package packit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Layers represents the set of layers managed by a buildpack.
type Layers struct {
	// Path is the absolute location of the set of layers managed by a buildpack
	// on disk.
	Path string
}

// Get will either create a new layer with the given name and layer types. If a
// layer already exists on disk, then the layer metadata will be retrieved from
// disk and returned instead.
func (l Layers) Get(name string, layerTypes ...LayerType) (Layer, error) {
	layer := Layer{
		Path:      filepath.Join(l.Path, name),
		Name:      name,
		SharedEnv: Environment{},
		BuildEnv:  Environment{},
		LaunchEnv: Environment{},
	}

	_, err := toml.DecodeFile(filepath.Join(l.Path, fmt.Sprintf("%s.toml", name)), &layer)
	if err != nil {
		if !os.IsNotExist(err) {
			return Layer{}, fmt.Errorf("failed to parse layer content metadata: %s", err)
		}
	}

	layer.SharedEnv, err = newEnvironmentFromPath(filepath.Join(l.Path, name, "env"))
	if err != nil {
		return Layer{}, err
	}

	layer.BuildEnv, err = newEnvironmentFromPath(filepath.Join(l.Path, name, "env.build"))
	if err != nil {
		return Layer{}, err
	}

	layer.LaunchEnv, err = newEnvironmentFromPath(filepath.Join(l.Path, name, "env.launch"))
	if err != nil {
		return Layer{}, err
	}

	for _, layerType := range layerTypes {
		switch layerType {
		case BuildLayer:
			layer.Build = true
		case CacheLayer:
			layer.Cache = true
		case LaunchLayer:
			layer.Launch = true
		}
	}

	return layer, nil
}
