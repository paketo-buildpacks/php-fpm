# PHP FPM Cloud Native Buildpack
A Cloud Native Buildpack for configuring FPM (FastCGI Process Manager) in PHP apps.

The buildpack generates the FPM configuration file
with the minimal set of options to get FPM to work, and incorporates
configuration snippets from different sources. The final FPM configuration file
is available at
`/layers/paketo-buildpacks_php-fpm/php-fpm-config/php-fpm-config/base.conf`, or
locatable through the buildpack-set `$PHP_FPM_PATH` environment variable.

## FPM Configuration Sources
The base configuration file generated in this buildpack includes some default
configuration, and `include` snippets from four different sources. The
configuration from the different inclusions are considered in the following
order of precedence (lowest to highest), with highest precedence configuration
options overriding conflicting settings from lower precendence sources.

1. (Lowest precendence) PHP Distribution default settings, which are taken from
   `$PHPRC/php-fpm.d/www.conf.default` (set by default in the PHP Dist
   buildpack).
2. Configuration set in this buildpack (see config/php-fpm-buildpack.conf)
3. Configuration set by other buildpacks, located at
   `/workspace/.php.fpm.bp/*.conf`
4. (Highest precendence) Configuration set by the user, located in the app
   source directory under `.php.fpm.d/*.conf`

## Integration

The PHP FPM CNB provides `php-fpm`, which can be required by subsequent
buildpacks. In order to configure FPM, the buildpack requires the `php`
dependency at build-time and run-time, which can be provided by a buildpack
like [Paketo PHP Dist](https://github.com/paketo-buildpacks/php-dist).

## Usage

To package this buildpack for consumption:

```
$ ./scripts/package.sh
```

This builds the buildpack's Go source using `GOOS=linux` by default. You can
supply another value as the first argument to `package.sh`.

## Run Tests

To run all unit tests, run:
```
./scripts/unit.sh
```

To run all integration tests, run:
```
./scripts/integration.sh
```

## Debug Logs
For extra debug logs from the image build process, set the `$BP_LOG_LEVEL`
environment variable to `DEBUG` at build-time (ex. `pack build my-app --env
BP_LOG_LEVEL=DEBUG` or through a  [`project.toml`
file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md).
