package paketosbom_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPaketoSBOM(t *testing.T) {
	suite := spec.New("paketosbom", spec.Report(report.Terminal{}))
	suite("sbom", testPaketoSBOM)
	suite.Run(t)
}
