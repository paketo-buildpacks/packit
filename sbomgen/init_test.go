package sbomgen_test

import (
	"testing"
	"time"

	"github.com/onsi/gomega/format"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitSBOM(t *testing.T) {
	format.MaxLength = 0

	suite := spec.New("sbomgen", spec.Report(report.Terminal{}))
	suite("Formats", testFormats)
	suite("SyftCLIScanner", testSyftCLIScanner)
	suite.Run(t)
}

type externalRef struct {
	Category string `json:"referenceCategory"`
	Locator  string `json:"referenceLocator"`
	Type     string `json:"referenceType"`
}

type pkg struct {
	ExternalRefs     []externalRef `json:"externalRefs"`
	LicenseConcluded string        `json:"licenseConcluded"`
	LicenseDeclared  string        `json:"licenseDeclared"`
	Name             string        `json:"name"`
	Version          string        `json:"versionInfo"`
}

type spdxOutput struct {
	Packages          []pkg  `json:"packages"`
	SPDXVersion       string `json:"spdxVersion"`
	DocumentNamespace string `json:"documentNamespace"`
	CreationInfo      struct {
		Created time.Time `json:"created"`
	} `json:"creationInfo"`
}
