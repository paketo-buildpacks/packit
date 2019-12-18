package packit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type LayerType uint8

const (
	BuildLayer LayerType = iota
	CacheLayer
	LaunchLayer
)

type Layer struct {
	Path      string                 `toml:"-"`
	Name      string                 `toml:"-"`
	Build     bool                   `toml:"build"`
	Launch    bool                   `toml:"launch"`
	Cache     bool                   `toml:"cache"`
	SharedEnv Environment            `toml:"-"`
	BuildEnv  Environment            `toml:"-"`
	LaunchEnv Environment            `toml:"-"`
	Metadata  map[string]interface{} `toml:"metadata"`
}

type Layers struct {
	Path string
}

func (l Layers) Get(name string, layerTypes ...LayerType) (Layer, error) {
	layer := Layer{
		Path:      filepath.Join(l.Path, name),
		Name:      name,
		SharedEnv: NewEnvironment(),
		BuildEnv:  NewEnvironment(),
		LaunchEnv: NewEnvironment(),
	}

	_, err := toml.DecodeFile(filepath.Join(l.Path, fmt.Sprintf("%s.toml", name)), &layer)
	if err != nil {
		if !os.IsNotExist(err) {
			return Layer{}, fmt.Errorf("failed to parse layer content metadata: %s", err)
		}
	}

	layer.SharedEnv, err = NewEnvironmentFromPath(filepath.Join(l.Path, name, "env"))
	if err != nil {
		return Layer{}, err
	}

	layer.BuildEnv, err = NewEnvironmentFromPath(filepath.Join(l.Path, name, "env.build"))
	if err != nil {
		return Layer{}, err
	}

	layer.LaunchEnv, err = NewEnvironmentFromPath(filepath.Join(l.Path, name, "env.launch"))
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

	err = os.MkdirAll(layer.Path, os.ModePerm)
	if err != nil {
		return Layer{}, err
	}

	return layer, nil
}
