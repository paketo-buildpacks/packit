package sbom_test

import (
	"testing"

	"github.com/onsi/gomega/format"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestSBOM(t *testing.T) {
	format.MaxLength = 0

	suite := spec.New("sbom", spec.Report(report.Terminal{}))
	suite("CycloneDXFormatter", testCycloneDXFormatter)
	suite("SBOM", testSBOM)
	suite("SPDXFormatter", testSPDXFormatter)
	suite("SyftFormatter", testSyftFormatter)
	suite("Entries", testEntries)
	suite.Run(t)
}
