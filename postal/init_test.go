package postal_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPostal(t *testing.T) {
	suite := spec.New("packit/postal", spec.Report(report.Terminal{}))
	suite("Service", testService)
	suite.Run(t)
}
