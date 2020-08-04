package packit

// OptionConfig is the set of configurable options for the Build and Detect
// functions.
type OptionConfig struct {
	exitHandler ExitHandler
	args        []string
	tomlWriter  TOMLWriter
	envWriter   EnvironmentWriter
}

// Option declares a function signature that can be used to define optional
// modifications to the behavior of the Detect and Build functions.
type Option func(config OptionConfig) OptionConfig

//go:generate faux --interface ExitHandler --output fakes/exit_handler.go

// ExitHandler serves as the interface for types that can handle an error
// during the Detect or Build functions. ExitHandlers are responsible for
// translating error values into exit codes according the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#detection and
// https://github.com/buildpacks/spec/blob/main/buildpack.md#build.
type ExitHandler interface {
	Error(error)
}

// TOMLWriter serves as the interface for types that can handle the writing of
// TOML files. TOMLWriters take a path to a file location on disk and a
// datastructure to marshal.
type TOMLWriter interface {
	Write(path string, value interface{}) error
}

// EnvironmentWriter serves as the interface for types that can write an
// Environment to a directory on disk according to the specification:
// https://github.com/buildpacks/spec/blob/main/buildpack.md#provided-by-the-buildpacks.
type EnvironmentWriter interface {
	Write(dir string, env map[string]string) error
}

// WithExitHandler is an Option that overrides the ExitHandler for a given
// invocation of Build or Detect.
func WithExitHandler(exitHandler ExitHandler) Option {
	return func(config OptionConfig) OptionConfig {
		config.exitHandler = exitHandler
		return config
	}
}

// WithArgs is an Option that overrides the value of os.Args for a given
// invocation of Build or Detect.
func WithArgs(args []string) Option {
	return func(config OptionConfig) OptionConfig {
		config.args = args
		return config
	}
}
