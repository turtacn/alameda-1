package main

import (
	"github.com/containers-ai/alameda/admission-controller/cmd/app"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "admission-controller",
	Short: "alameda admission-controller",
	Long:  "",
}

func init() {
	RootCmd.AddCommand(app.RunCmd)
	RootCmd.AddCommand(app.VersionCmd)
}

func main() {
	RootCmd.Execute()
}
