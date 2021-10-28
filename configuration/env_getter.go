package configuration

import (
	"fmt"
	"os"
	"strconv"

	"github.com/mattn/go-shellwords"
)

type EnvGetter struct {
}

func NewEnvGetter() EnvGetter {
	return EnvGetter{}
}

func (e EnvGetter) LookupEnv(name string) (string, bool) {
	return os.LookupEnv(name)
}

func (e EnvGetter) LookupEnvWithDefault(name string, defaultVal string) string {
	if s, ok := os.LookupEnv(name); ok {
		return s
	}
	return defaultVal
}

func (e EnvGetter) GetEnvAsBool(name string) bool {
	s, ok := os.LookupEnv(name)
	if !ok {
		return false
	}

	if s == "" {
		return true // A boolean env var is true if it's set in the env with no value
	}

	t, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}

	return t
}

func (e EnvGetter) GetEnvAsShellWords(name string) (words []string, err error) {
	shellwordsParser := shellwords.NewParser()
	shellwordsParser.ParseEnv = true

	if raw, ok := os.LookupEnv(name); ok {
		words, err = shellwordsParser.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse value of '%s' as shell words: %w", name, err)
		}
	}
	return words, nil
}
