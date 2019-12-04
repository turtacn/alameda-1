package app

import (
	"github.com/spf13/cobra"
)

const (
	envVarPrefix = "ALAMEDA_RABBITMQ"
)

var (
	RootCmd = &cobra.Command{
		Use:   "rabbitmq",
		Short: "rabbitmq publish",
		Long:  "",
	}
)

func init() {
	RootCmd.AddCommand(PublishCmd)
}
