package cargo

import (
	"io"

	"github.com/cloudfoundry/packit/pexec"
	"github.com/cloudfoundry/packit/scribe"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(execution pexec.Execution) error
}

type PrePackager struct {
	executable Executable
	logger     scribe.Logger
	output     io.Writer
}

func NewPrePackager(executable Executable, logger scribe.Logger, output io.Writer) PrePackager {
	return PrePackager{
		executable: executable,
		logger:     logger,
		output:     output,
	}
}

func (p PrePackager) Execute(scriptPath, rootDir string) error {
	if scriptPath == "" {
		return nil
	}

	p.logger.Process("Executing pre-packaging script: %s", scriptPath)

	err := p.executable.Execute(pexec.Execution{
		Args:   []string{"-c", scriptPath},
		Dir:    rootDir,
		Stdout: p.output,
		Stderr: p.output,
	})
	if err != nil {
		return err
	}

	p.logger.Break()
	return nil
}
