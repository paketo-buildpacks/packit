package servicebindings_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitServiceBindings(t *testing.T) {
	suite := spec.New("packit/servicebindings", spec.Report(report.Terminal{}))
	suite("Resolver", testResolver)
	suite("Entry", testEntry)
	suite.Run(t)
}
