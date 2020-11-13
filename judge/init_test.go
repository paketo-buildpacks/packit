package judge_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPostal(t *testing.T) {
	suite := spec.New("packit/judge", spec.Report(report.Terminal{}))
	suite("PlanEntryHandler", testPlanEntryHandler)
	suite.Run(t)
}
