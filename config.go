package phpfpm

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2/fs"
)

type PhpFpmConfig struct {
	PhpDistribution string
	PhpFpmBuildpack string
	OtherBuildpacks string
	UserInclude     string
}

type Config struct{}

func NewConfig() Config {
	return Config{}
}

// Write takes in various source paths of FPM configurations, and uses the
// expected configuration path locations as `include` snippest in the base
// config file which will be located inside the PHP FPM layer. The base config
// file is what will be used by the FPM command.
func (c Config) Write(phpFpmLayer, phpDistPath, workingDir, cnbPath string) (string, error) {
	tmpl, err := template.New("php-fpm-base.conf").ParseFiles(filepath.Join(cnbPath, "config", "php-fpm-base.conf"))
	if err != nil {
		return "", fmt.Errorf("failed to parse FPM config template: %w", err)
	}

	// Configuration set by this buildpack
	fpmBuildpackConfig := filepath.Join(phpFpmLayer, "buildpack.conf")
	err = fs.Copy(filepath.Join(cnbPath, "config", "php-fpm-buildpack.conf"), fpmBuildpackConfig)
	if err != nil {
		return "", fmt.Errorf("failed to copy buildpack FPM config: %w", err)
	}

	// If the PHP Dist path is empty ($PHPRC is not set), don't include it in the
	// config for clarity
	distPath := filepath.Join(phpDistPath, "php-fpm.d", "www.conf.default")
	if phpDistPath == "" {
		distPath = ""
	}

	// Create the directory for subsequent buildpacks to include configuration.
	// FPM will fail to start up if configuration includes a snippet in a
	// non-existent directory.
	bpPath := filepath.Join(workingDir, ".php.fpm.bp")
	err = os.MkdirAll(bpPath, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create %s: %w", bpPath, err)
	}

	// If there's a user-provided FPM conf, include it in the base configuration.
	userPath := filepath.Join(workingDir, ".php.fpm.d", "*.conf")
	_, err = os.Stat(filepath.Join(workingDir, ".php.fpm.d"))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			// untested
			return "", fmt.Errorf("failed to stat %s/.php.fpm.d: %w", workingDir, err)
		}
		userPath = ""
	}

	// The FPM configuration will be considered in the following order of precedence (from lowest to highest):
	// 1. The PHP Distribution default settings (will be omitted entirely if nonexistent)
	// 2. The buildpack-set configurations (see config/php-fpm-buildpack.conf)
	// 3. Configuration files contributed by other buildpacks (likely depending on the webserver of choice)
	// 4. User included configuration files (will be omitted entirely if nonexistent)
	data := PhpFpmConfig{
		PhpDistribution: distPath,
		PhpFpmBuildpack: fpmBuildpackConfig,
		OtherBuildpacks: filepath.Join(workingDir, ".php.fpm.bp", "*.conf"),
		UserInclude:     userPath,
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, data)
	if err != nil {
		// not tested
		return "", err
	}

	f, err := os.OpenFile(filepath.Join(phpFpmLayer, "base.conf"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, &b)
	if err != nil {
		// not tested
		return "", err
	}

	return f.Name(), nil
}
