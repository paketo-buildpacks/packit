package biome_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitBiome(t *testing.T) {
	suite := spec.New("packit/biome", spec.Report(report.Terminal{}))
	suite("Biome", testBiome)
	suite.Run(t)
}
