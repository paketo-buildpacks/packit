package chronos_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitChronos(t *testing.T) {
	suite := spec.New("packit/chronos", spec.Report(report.Terminal{}))
	suite("Clock", testClock)
	suite.Run(t)
}
