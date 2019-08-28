package keycodes

import (
	"github.com/containers-ai/alameda/datahub/tools/license-utils/pkg/keycodes"
	"github.com/spf13/cobra"
)

var ActivateCmd = &cobra.Command{
	Use:   "activate",
	Short: "activate keycode signature Data",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		keycodes.Activate(filePath)
	},
}

func init() {
	parseActivateFlag()
}

func parseActivateFlag() {
	ActivateCmd.Flags().StringVar(&filePath, "path", "", "The file path of keycode signature data.")
}
