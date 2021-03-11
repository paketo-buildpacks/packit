package pexec_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	existingPath string
	fakeCLI      string
)

func TestUnitExec(t *testing.T) {
	var Expect = NewWithT(t).Expect

	suite := spec.New("packit/pexec", spec.Report(report.Terminal{}))
	suite("pexec", testPexec)

	var err error
	fakeCLI, err = gexec.Build("github.com/paketo-buildpacks/packit/fakes/some-executable")
	Expect(err).NotTo(HaveOccurred())

	existingPath = os.Getenv("PATH")
	os.Setenv("PATH", filepath.Dir(fakeCLI))

	t.Cleanup(func() {
		os.Setenv("PATH", existingPath)
		gexec.CleanupBuildArtifacts()
	})

	suite.Run(t)
}
