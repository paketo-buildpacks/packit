package planning_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitBuildplan(t *testing.T) {
	suite := spec.New("buildplan", spec.Report(report.Terminal{}))
	suite("Metadata", testMetadata)
	suite.Run(t)
}
