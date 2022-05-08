package reload_test

import (
	"testing"

	"github.com/onsi/gomega/format"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestReload(t *testing.T) {
	format.MaxLength = 0

	suite := spec.New("reload", spec.Report(report.Terminal{}))
	suite("Reload", testReload)
	suite.Run(t)
}
