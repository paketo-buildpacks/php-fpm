package phpfpm_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPhpFpm(t *testing.T) {
	suite := spec.New("php-fpm", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Build", testBuild, spec.Sequential())
	suite("Detect", testDetect)
	suite("Config", testConfig)
	suite.Run(t)
}
