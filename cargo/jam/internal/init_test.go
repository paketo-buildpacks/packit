package internal_test

import (
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitCargo(t *testing.T) {
	gomega.SetDefaultEventuallyTimeout(10 * time.Second)

	suite := spec.New("cargo/jam/internal", spec.Report(report.Terminal{}))
	suite("BuildpackInspector", testBuildpackInspector)
	suite("Formatter", testFormatter)
	suite.Run(t)
}
