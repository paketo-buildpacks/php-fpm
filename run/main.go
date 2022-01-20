package main

import (
	"github.com/paketo-buildpacks/packit/v2"
	phpfpm "github.com/paketo-buildpacks/php-fpm"
)

func main() {
	packit.Run(phpfpm.Detect(), phpfpm.Build())
}
