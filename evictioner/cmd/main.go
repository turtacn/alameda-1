package main

import (
	"github.com/containers-ai/alameda/cmd/app"
	evictioner_app "github.com/containers-ai/alameda/evictioner/cmd/app"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "evictioner",
	Short: "alameda evictioner",
	Long:  "",
}

var (
	// VERSION is sofeware version
	VERSION string
	// BUILD_TIME is build time
	BUILD_TIME string
	// GO_VERSION is go version
	GO_VERSION string
)

func init() {
	RootCmd.AddCommand(evictioner_app.RunCmd)
	RootCmd.AddCommand(app.VersionCmd)
}

func main() {
	setSoftwareInfo()
	RootCmd.Execute()
}

func setSoftwareInfo() {
	app.VERSION = VERSION
	app.BUILD_TIME = BUILD_TIME
	app.GO_VERSION = GO_VERSION
	app.PRODUCT_NAME = "evictioner"
}
