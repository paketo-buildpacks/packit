package packit

import (
	"fmt"
	"os"
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

func (l *Layer) Reset() error {
	l.SharedEnv = NewEnvironment()
	l.BuildEnv = NewEnvironment()
	l.LaunchEnv = NewEnvironment()
	l.Metadata = nil

	err := os.RemoveAll(l.Path)
	if err != nil {
		return fmt.Errorf("error could not remove file: %s", err)
	}

	err = os.MkdirAll(l.Path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error could not create directory: %s", err)
	}

	return nil
}
