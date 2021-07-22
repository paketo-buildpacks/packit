package extensions_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitExtensions(t *testing.T) {
	suite := spec.New("packit/extensions", spec.Report(report.Terminal{}))
	suite("ServiceBindingsManager", testServiceBindingsManager)
	suite.Run(t)
}
