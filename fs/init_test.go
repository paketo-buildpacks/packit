package fs_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitFS(t *testing.T) {
	suite := spec.New("packit/fs", spec.Report(report.Terminal{}))
	suite("Move", testMove)
	suite("IsEmptyDir", testIsEmptyDir)
	suite("ChecksumCalculator", testChecksumCalculator)
	suite.Run(t)
}
