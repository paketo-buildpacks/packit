package configuration_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitConfig(t *testing.T) {
	suite := spec.New("packit/configuration", spec.Report(report.Terminal{}))
	suite("EnvGetter", testEnvGetter, spec.Sequential())
	suite.Run(t)
}
