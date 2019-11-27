package exec_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitExec(t *testing.T) {
	suite := spec.New("packit/exec", spec.Report(report.Terminal{}))
	suite("Exec", testExec)
	suite.Run(t)
}
