package draft_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDraft(t *testing.T) {
	suite := spec.New("packit/draft", spec.Report(report.Terminal{}))
	suite("Planner", testPlanner)
	suite.Run(t)
}
