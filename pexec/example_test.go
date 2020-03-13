package pexec_test

import (
	"os"

	"github.com/cloudfoundry/packit/pexec"
)

func ExampleExecute() {
	ls := pexec.NewExecutable("ls")

	err := ls.Execute(pexec.Execution{
		Args:   []string{"-al", "/"},
		Stdout: os.Stdout,
	})
	if err != nil {
		panic(err)
	}

	err = ls.Execute(pexec.Execution{
		Args:   []string{"-R", "/etc"},
		Stdout: os.Stdout,
	})
	if err != nil {
		panic(err)
	}
}
