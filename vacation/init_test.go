package vacation_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestVacation(t *testing.T) {
	suite := spec.New("vacation", spec.Report(report.Terminal{}))
	suite("VacationArchive", testVacationArchive)
	suite("VacationTar", testVacationTar)
	suite("VacationTarGzip", testVacationTarGzip)
	suite("VacationTarXZ", testVacationTarXZ)
	suite("VacationText", testVacationText)
	suite("VacationZip", testVacationZip)
	suite.Run(t)
}
