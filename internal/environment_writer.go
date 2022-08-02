package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type EnvironmentWriter struct{}

func NewEnvironmentWriter() EnvironmentWriter {
	return EnvironmentWriter{}
}

func (w EnvironmentWriter) Write(dir string, env map[string]string) error {
	if len(env) == 0 {
		return nil
	}

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	// this regex checks that map keys contain valid env var name characters,
	// per https://pubs.opengroup.org/onlinepubs/9699919799/
	validEnvVarRegex := regexp.MustCompile(`^[a-zA-Z_]{1,}[a-zA-Z0-9_]*$`)

	for key, value := range env {
		parts := strings.SplitN(key, ".", 2)
		if !validEnvVarRegex.MatchString(parts[0]) {
			return fmt.Errorf("invalid environment variable name '%s'", parts[0])
		}
		err := os.WriteFile(filepath.Join(dir, key), []byte(value), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
