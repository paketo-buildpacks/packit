package sbom_test

import (
	"testing"

	"github.com/onsi/gomega/format"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitSBOM(t *testing.T) {
	format.MaxLength = 0

	suite := spec.New("sbom", spec.Report(report.Terminal{}))
	suite("Formatter", testFormatter)
	suite("FormattedReader", testFormattedReader)
	suite("SBOM", testSBOM)
	suite.Run(t)
}
