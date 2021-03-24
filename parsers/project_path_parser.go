package parsers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

//go:generate faux --interface ProjectPathParser --output fakes/project_path_parser.go
type ProjectPathParser interface {
	Get(path string, envVarName string) (projectPath string, err error)
}

// ProjectPathParser provides a mechanism for determining the proper working
// directory for the build process.
type projectPathParser struct{}

// NewProjectPathParser creates an instance of a ProjectPathParser.
func NewProjectPathParser() projectPathParser {
	return projectPathParser{}
}

// Get will resolve the environment variable. It
// validates that environment variable value is valid relative to the provided path.
func (p projectPathParser) Get(workingDirPath string, envVarName string) (string, error) {
	customProjPath := os.Getenv(envVarName)
	if customProjPath == "" {
		return "", nil
	}

	_, err := os.Stat(filepath.Join(workingDirPath, customProjPath))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("expected value derived from %s [%s] to be an existing directory", envVarName, customProjPath)
		}
		return "", err
	}
	return customProjPath, nil
}
