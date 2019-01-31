package main

import (
	"fmt"

	"github.com/containers-ai/alameda/datahub/cmd/app"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "datahub",
	Short: "alameda datahub",
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
	RootCmd.AddCommand(app.RunCmd)
}

func main() {
	printSoftwareInfo()
	RootCmd.Execute()
}

func printSoftwareInfo() {
	fmt.Println(fmt.Sprintf("Datahub Version: %s", VERSION))
	fmt.Println(fmt.Sprintf("Datahub Build Time: %s", BUILD_TIME))
	fmt.Println(fmt.Sprintf("Datahub GO Version: %s", GO_VERSION))
}
