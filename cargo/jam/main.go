package main

import (
	"fmt"
	"github.com/paketo-buildpacks/packit/cargo/jam/commands"
	"os"
)

func main() {
	err := commands.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to execute: %s", err)
		os.Exit(1)
	}
}
