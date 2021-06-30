package vacation_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestVacation(t *testing.T) {
	suite := spec.New("vacation", spec.Report(report.Terminal{}))
	suite("Archive", testArchive)
	suite("NopArchive", testNopArchive)
	suite("SymlinkSorting", testSymlinkSorting)
	suite("TarArchive", testTarArchive)
	suite("TarBzip2Archive", testTarBzip2Archive)
	suite("TarGzipArchive", testTarGzipArchive)
	suite("TarXZArchive", testTarXZArchive)
	suite("ZipArchive", testZipArchive)
	suite.Run(t)
}
