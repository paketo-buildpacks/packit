package commands_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestCommands(t *testing.T) {
	suite := spec.New("commands", spec.Report(report.Terminal{}))
	suite("CommandPopulator", testCommandPopulator)
	suite.Run(t)
}
