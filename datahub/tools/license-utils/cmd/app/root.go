package app

import (
	AppKeycodes "github.com/containers-ai/alameda/datahub/tools/license-utils/cmd/app/keycodes"
	Keycodes "github.com/containers-ai/alameda/datahub/tools/license-utils/pkg/keycodes"
	Setup "github.com/containers-ai/alameda/datahub/tools/license-utils/pkg/setup"
	"github.com/spf13/cobra"
)

const (
	DefaultDatahubAddress = "127.0.0.1:50050"
)

var RootCmd = &cobra.Command{
	Use:              "license-utils",
	Short:            "datahub keycode related utilities",
	Long:             "",
	TraverseChildren: true,
}

var (
	datahubAddress string
)

func init() {
	Keycodes.KeycodeInit(&datahubAddress)
	Setup.SetupInit(&datahubAddress)

	RootCmd.AddCommand(InteractiveCmd)
	RootCmd.AddCommand(AppKeycodes.KeycodeCmd)
	RootCmd.AddCommand(VersionCmd)
	RootCmd.PersistentFlags().StringVar(&datahubAddress, "address", DefaultDatahubAddress, "The address of datahub.")
}
