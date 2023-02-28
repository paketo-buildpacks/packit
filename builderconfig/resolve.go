package builderconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

const DefaultConfigPath = "/builder/config.toml"

type EnvLooker func(string) (string, bool)
type EnvSetter func(string, string) error
type EnvUnSetter func(string) error

type Config struct {
	path        string
	envSetter   EnvSetter
	envUnSetter EnvUnSetter
	envLooker   EnvLooker
}

var SupportedAPIs = []string{"0.1"}

// Option is a function for configuring a Config instance.
type Option func(config Config) Config

func WithPath(path string) Option {
	return func(config Config) Config {
		config.path = path
		return config
	}
}

func WithEnvLooker(looker EnvLooker) Option {
	return func(config Config) Config {
		config.envLooker = looker
		return config
	}
}

func WithEnvSetter(setter EnvSetter) Option {
	return func(config Config) Config {
		config.envSetter = setter
		return config
	}
}

func WithEnvUnSetter(unSetter EnvUnSetter) Option {
	return func(config Config) Config {
		config.envUnSetter = unSetter
		return config
	}
}

func New(options ...Option) Config {
	config := Config{
		path:        DefaultConfigPath,
		envSetter:   os.Setenv,
		envLooker:   os.LookupEnv,
		envUnSetter: os.Unsetenv,
	}
	for _, o := range options {
		config = o(config)
	}
	return config
}

type configFile struct {
	API   string      `toml:"api"`
	Build buildConfig `toml:"build"`
}

type buildConfig struct {
	Env []envVar `toml:"env"`
}

type envVar struct {
	Name      string `toml:"name"`
	Value     string `toml:"value"`
	Mode      string `toml:"mode"`
	Delimiter string `toml:"delim"`
}

func (c Config) Resolve() error {
	if _, err := os.Stat(c.path); err != nil {
		return nil
	}

	cf := configFile{}
	if _, err := toml.DecodeFile(c.path, &cf); err != nil {
		return fmt.Errorf("unable to parse builder config: %w", err)
	}

	if cf.API != "0.1" {
		return fmt.Errorf("invalid API for builder config: %s, supported APIs: %s", cf.API, SupportedAPIs)
	}

	for _, env := range cf.Build.Env {
		delim := env.Delimiter
		if delim == "" {
			delim = string(os.PathListSeparator)
		}
		switch env.Mode {
		case "", "default":
			if _, set := c.envLooker(env.Name); !set {
				if err := c.envSetter(env.Name, env.Value); err != nil {
					return fmt.Errorf("unable to parse config: %w", err)
				}
			}
		case "override":
			if err := c.envSetter(env.Name, env.Value); err != nil {
				return fmt.Errorf("unable to parse config: %w", err)
			}
		case "prepend":
			if value, set := c.envLooker(env.Name); !set {
				if err := c.envSetter(env.Name, env.Value); err != nil {
					return fmt.Errorf("unable to parse config: %w", err)
				}
			} else {
				if err := c.envSetter(env.Name, strings.Join([]string{env.Value, value}, delim)); err != nil {
					return fmt.Errorf("unable to parse config: %w", err)
				}
			}
		case "append":
			if value, set := c.envLooker(env.Name); !set {
				if err := c.envSetter(env.Name, env.Value); err != nil {
					return fmt.Errorf("unable to parse config: %w", err)
				}
			} else {
				if err := c.envSetter(env.Name, strings.Join([]string{value, env.Value}, delim)); err != nil {
					return fmt.Errorf("unable to parse config: %w", err)
				}
			}
		case "unset":
			if err := c.envUnSetter(env.Name); err != nil {
				return fmt.Errorf("unable to parse config: %w", err)
			}
		default:
			return fmt.Errorf("unknown mode: %s, supported modes: %s", env.Mode, []string{"default", "override", "append", "prepend", "unset"})
		}
	}
	return nil
}
