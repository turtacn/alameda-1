package main

import (
	"github.com/containers-ai/alameda/cmd/app"
	LicenseUtils "github.com/containers-ai/alameda/datahub/tools/license-utils/cmd/app"
)

var (
	// VERSION is software version
	VERSION string
	// BUILD_TIME is build time
	BUILD_TIME string
	// GO_VERSION is go version
	GO_VERSION string
)

func init() {
	setSoftwareInfo()
}

func setSoftwareInfo() {
	app.VERSION = VERSION
	app.BUILD_TIME = BUILD_TIME
	app.GO_VERSION = GO_VERSION
	app.PRODUCT_NAME = "license-utils"
}

func main() {
	LicenseUtils.RootCmd.Execute()
}
