package scribe_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitScribe(t *testing.T) {
	suite := spec.New("scribe", spec.Report(report.Terminal{}))
	suite("Bar", testBar)
	suite("Color", testColor)
	suite("FormattedList", testFormattedList)
	suite("FormattedMap", testFormattedMap)
	suite("Log", testLog)
	suite.Run(t)
}
