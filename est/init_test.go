package est_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPackit(t *testing.T) {
	suite := spec.New("Est", spec.Report(report.Terminal{}))
	suite("FileExists", testFileExists)
	suite.Run(t)
}
