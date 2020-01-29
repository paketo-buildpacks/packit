package vacation_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestVacation(t *testing.T) {
	suite := spec.New("vacation", spec.Report(report.Terminal{}))
	suite("Vacation", testVacation)
	suite.Run(t)
}
