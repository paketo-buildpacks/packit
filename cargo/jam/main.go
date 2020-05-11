package main

import (
	"fmt"
	"os"

	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/cargo/jam/commands"
	"github.com/paketo-buildpacks/packit/cargo/jam/internal"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
)

type Command interface {
	Execute(args []string) error
}

func main() {
	if len(os.Args) < 2 {
		fail("missing command")
	}

	var command Command

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
		command = commands.NewPack(directoryDuplicator, buildpackParser, prePackager, dependencyCacher, fileBundler, tarBuilder, os.Stdout)

	case "summarize":
		inspector := internal.NewBuildpackInspector()
		formatter := internal.NewFormatter(os.Stdout)
		command = commands.NewSummarize(inspector, formatter)

	default:
		fail("unknown command: %q", os.Args[1])
	}

	if err := command.Execute(os.Args[2:]); err != nil {
		fail("failed to execute: %s", err)
	}

}

func fail(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}
