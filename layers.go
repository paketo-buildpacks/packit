package packit

import (
	"os"
	"path/filepath"
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

	err := os.MkdirAll(layer.Path, os.ModePerm)
	if err != nil {
		return Layer{}, err
	}

	return layer, nil
}
