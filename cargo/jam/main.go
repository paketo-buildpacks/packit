package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/cargo/jam/commands"
	"github.com/cloudfoundry/packit/pexec"
	"github.com/cloudfoundry/packit/scribe"
)

func main() {
	if len(os.Args) < 2 {
		fail("missing command")
	}

	switch os.Args[1] {
	case "pack":
		logger := scribe.NewLogger(os.Stdout)
		bash := pexec.NewExecutable("bash")

		transport := cargo.NewTransport()
		directoryDuplicator := cargo.NewDirectoryDuplicator()
		buildpackParser := cargo.NewBuildpackParser()
		fileBundler := cargo.NewFileBundler()
		tarBuilder := cargo.NewTarBuilder(logger)
		prePackager := cargo.NewPrePackager(bash, logger, scribe.NewWriter(os.Stdout, scribe.WithIndent(2)))
		dependencyCacher := cargo.NewDependencyCacher(transport, logger)
		command := commands.NewPack(directoryDuplicator, buildpackParser, prePackager, dependencyCacher, fileBundler, tarBuilder, os.Stdout)

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
