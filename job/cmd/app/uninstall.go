package app

import (
	"github.com/spf13/cobra"
)

var (
	UninstallCmd = &cobra.Command{
		Use:   "uninstall",
		Short: "uninstall prometheus rules",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			scope.Info("uninstall")
		},
	}
)
