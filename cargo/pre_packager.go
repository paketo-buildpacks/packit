package cargo

import (
	"github.com/cloudfoundry/packit/pexec"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(execution pexec.Execution) (stdOut string, stdError string, err error)
}

type PrePackager struct {
	executable Executable
}

func NewPrePackager(executable Executable) PrePackager {
	return PrePackager{
		executable: executable,
	}
}

func (p PrePackager) Execute(scriptPath, rootDir string) error {
	if scriptPath == "" {
		return nil
	}
	_, _, err := p.executable.Execute(pexec.Execution{
		Args: []string{"-c", scriptPath},
		Dir:  rootDir,
	})
	return err
}
