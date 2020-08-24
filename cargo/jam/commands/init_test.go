package commands_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitCommands(t *testing.T) {
	suite := spec.New("jam/commands", spec.Report(report.Terminal{}))
	suite("Pack", testPack)
	suite("Summarize", testSummarize)
	suite.Run(t)
}
