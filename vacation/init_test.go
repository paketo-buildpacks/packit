package vacation_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestVacation(t *testing.T) {
	suite := spec.New("vacation", spec.Report(report.Terminal{}))
	suite("Archive", testArchive)
	suite("Bzip2Archive", testBzip2Archive)
	suite("Executable", testExecutable)
	suite("GzipArchive", testGzipArchive)
	suite("LinkSorting", testLinkSorting)
	suite("NopArchive", testNopArchive)
	suite("TarArchive", testTarArchive)
	suite("XZArchive", testXZArchive)
	suite("ZipArchive", testZipArchive)
	suite.Run(t)
}
