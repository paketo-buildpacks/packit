package main

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/lager"

	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/cargo/jam/commands"
	"github.com/cloudfoundry/packit/pexec"
)

func main() {
	if len(os.Args) < 2 {
		fail("missing command")
	}

	switch os.Args[1] {
	case "pack":
		directoryDuplicator := cargo.NewDirectoryDuplicator()
		buildpackParser := cargo.NewBuildpackParser()
		fileBundler := cargo.NewFileBundler()
		tarBuilder := cargo.NewTarBuilder()
		prePackager := cargo.NewPrePackager(pexec.NewExecutable("bash", lager.NewLogger("pre-packager")))
		command := commands.NewPack(directoryDuplicator, buildpackParser, prePackager, fileBundler, tarBuilder, os.Stdout)

		if err := command.Execute(os.Args[2:]); err != nil {
			fail("failed to execute pack command: %s", err)
		}

	default:
		fail("unknown command: %q", os.Args[1])
	}
}

func fail(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}
