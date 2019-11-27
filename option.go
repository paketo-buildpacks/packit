package packit

type Config struct {
	exitHandler ExitHandler
	args        []string
	tomlWriter  TOMLWriter
	envWriter   EnvironmentWriter
}

type Option func(config Config) Config

//go:generate faux --interface ExitHandler --output fakes/exit_handler.go
type ExitHandler interface {
	Error(error)
}

type TOMLWriter interface {
	Write(path string, value interface{}) error
}

type EnvironmentWriter interface {
	Write(dir string, env map[string]string) error
}

func WithExitHandler(exitHandler ExitHandler) Option {
	return func(config Config) Config {
		config.exitHandler = exitHandler
		return config
	}
}

func WithArgs(args []string) Option {
	return func(config Config) Config {
		config.args = args
		return config
	}
}
