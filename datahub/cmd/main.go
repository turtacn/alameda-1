package main

import (
	"github.com/containers-ai/alameda/datahub/cmd/app"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "datahub",
	Short: "alameda datahub",
	Long:  "",
}

func init() {
	RootCmd.AddCommand(app.RunCmd)
}

func main() {
	RootCmd.Execute()
}
