package app

import (
	"github.com/containers-ai/alameda/cmd/app"
	"github.com/spf13/cobra"
)

var (
	VersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display the datahub license-utils version",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			app.PrintSoftwareVer()
		},
	}
)
