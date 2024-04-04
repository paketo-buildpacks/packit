package internal_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPostalInternal(t *testing.T) {
	suite := spec.New("packit/postal/internal", spec.Report(report.Terminal{}))
	suite("DependencyMappings", testDependencyMappings)
	suite("DependencyMirror", testDependencyMirror)

	suite.Run(t)
}
