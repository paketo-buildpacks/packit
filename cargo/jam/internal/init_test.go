package internal_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitCargo(t *testing.T) {
	suite := spec.New("cargo/jam/internal", spec.Report(report.Terminal{}))
	suite("BuildpackInspector", testBuildpackInspector)
	suite("Formatter", testFormatter)
	suite.Run(t)
}
