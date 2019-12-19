package packit_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPackit(t *testing.T) {
	suite := spec.New("packit", spec.Report(report.Terminal{}))
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite("Environment", testEnvironment)
	suite("Layer", testLayer)
	suite("Layers", testLayers)
	suite.Run(t)
}
