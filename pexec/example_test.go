package pexec_test

import (
	"os"

	"github.com/cloudfoundry/packit/pexec"
)

func ExampleExecute() {
	echo := pexec.NewExecutable("echo")

	err := echo.Execute(pexec.Execution{
		Args:   []string{"hello from pexec"},
		Stdout: os.Stdout,
	})
	if err != nil {
		panic(err)
	}

	// Output: hello from pexec
}
