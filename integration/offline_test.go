package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testOffline(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run in an offline environment", func() {
		var (
			image      occam.Image
			container  occam.Container
			container2 occam.Container

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Container.Remove.Execute(container2.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("generates a functional php-fpm config file and makes it available at build and launch", func() {
			var (
				logs fmt.Stringer
				err  error
			)

			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					phpOfflineBuildpack,
					offlineBuildpack,
					buildPlanBuildpack,
				).
				WithNetwork("none").
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			container, err = docker.Container.Run.
				WithCommand(fmt.Sprintf("cat /layers/%s/php-fpm-config/base.conf", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))).
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(
				And(
					ContainSubstring("include = /layers/paketo-buildpacks_php-dist/php/etc/php-fpm.d/www.conf.default"),
					ContainSubstring(fmt.Sprintf("include = /layers/%s/php-fpm-config/buildpack.conf", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))),
					ContainSubstring("include = /workspace/.php.fpm.bp/*.conf"),
					ContainSubstring("include = /workspace/.php.fpm.d/*.conf"),
				),
			)

			container2, err = docker.Container.Run.
				WithCommand(fmt.Sprintf("php-fpm -y /layers/%s/php-fpm-config/base.conf", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))).
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(container2.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(
				And(
					ContainSubstring("NOTICE: fpm is running, pid 1"),
					ContainSubstring("NOTICE: ready to handle connections"),
				),
			)

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Setting up the FPM configuration file",
				"  Configuring build environment",
				MatchRegexp(fmt.Sprintf(`    PHP_FPM_PATH -> "/layers/%s/php-fpm-config/base.conf"`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))),
				"",
				"  Configuring launch environment",
				MatchRegexp(fmt.Sprintf(`    PHP_FPM_PATH -> "/layers/%s/php-fpm-config/base.conf"`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))),
			))
		})
	})
}
