package parsers_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestParsers(t *testing.T) {
	suite := spec.New("utility", spec.Report(report.Terminal{}))
	suite("ProjectPathParser", testProjectPathParser)
	suite.Run(t)
}
