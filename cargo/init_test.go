package cargo_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitCargo(t *testing.T) {
	suite := spec.New("cargo", spec.Report(report.Terminal{}))
	suite("Transport", testTransport)
	suite.Run(t)
}
