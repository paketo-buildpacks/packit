package cargo_test

import (
	"errors"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitCargo(t *testing.T) {
	suite := spec.New("cargo", spec.Report(report.Terminal{}))
	suite("BuildpackParser", testBuildpackParser)
	suite("ExtensionParser", testExtensionParser)
	suite("Config", testConfig)
	suite("ExtensionConfig", testExtensionConfig)
	suite("DirectoryDuplicator", testDirectoryDuplicator)
	suite("Transport", testTransport)
	suite("ValidatedReader", testValidatedReader)
	suite("Checksum", testChecksum)
	suite.Run(t)
}

type errorReader struct{}

func (r errorReader) Read(p []byte) (int, error) {
	return 0, errors.New("failed to read")
}
