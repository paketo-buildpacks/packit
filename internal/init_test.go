package internal_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitInternal(t *testing.T) {
	suite := spec.New("packit/internal", spec.Report(report.Terminal{}))
	suite("EnvironmentWriter", testEnvironmentWriter)
	suite("ExitHandler", testExitHandler)
	suite("TOMLWriter", testTOMLWriter)
	suite.Run(t)
}
