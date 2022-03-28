package phpfpm

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go

// EntryResolver defines the interface for picking the most relevant entry from
// the Buildpack Plan entries.
type EntryResolver interface {
	MergeLayerTypes(name string, entries []packit.BuildpackPlanEntry) (launch, build bool)
}

//go:generate faux --interface ConfigWriter --output fakes/config_writer.go

// ConfigWriter puts together the snippets of the php-fpm.conf file from
// different sources via Go templating.
type ConfigWriter interface {
	Write(layer, phpDistPath, workingDir, cnbPath string) (string, error)
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
// Build will create a layer dedicated to PHP FPM configuration, configure default FPM
// settings, incorporate other configuration sources, and make the
// configuration available at both build-time and
// launch-time.
func Build(entryResolver EntryResolver, config ConfigWriter, clock chronos.Clock, logger scribe.Emitter) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		logger.Debug.Process("Getting the layer associated with FPM")
		configLayer, err := context.Layers.Get(PhpFpmConfigLayerName)
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Debug.Subprocess(configLayer.Path)
		logger.Debug.Break()

		configLayer, err = configLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Process("Setting up the FPM configuration file")

		logger.Debug.Subprocess("Getting the PHP Distribution $PHPRC path")
		phpDistPath, ok := os.LookupEnv("PHPRC")
		if !ok {
			logger.Debug.Subprocess("The $PHPRC isn't set")
		}
		logger.Debug.Subprocess("PHPRC: %s", phpDistPath)
		logger.Debug.Break()

		phpFpmPath, err := config.Write(configLayer.Path, phpDistPath, context.WorkingDir, context.CNBPath)
		if err != nil {
			return packit.BuildResult{}, err
		}

		launch, build := entryResolver.MergeLayerTypes(PhpFpmDependency, context.Plan.Entries)
		configLayer.Launch = launch
		configLayer.Build = build

		configLayer.SharedEnv.Default("PHP_FPM_PATH", phpFpmPath)
		logger.EnvironmentVariables(configLayer)

		return packit.BuildResult{
			Layers: []packit.Layer{configLayer},
		}, nil
	}
}
