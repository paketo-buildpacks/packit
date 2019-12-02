package pexec_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitExec(t *testing.T) {
	suite := spec.New("packit/pexec", spec.Report(report.Terminal{}))
	suite("pexec", testPexec)
	suite.Run(t)
}
