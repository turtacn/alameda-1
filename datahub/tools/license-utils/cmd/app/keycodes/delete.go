package keycodes

import (
	"github.com/containers-ai/alameda/datahub/tools/license-utils/pkg/keycodes"
	"github.com/spf13/cobra"
)

var DeleteKeycodeCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete keycode",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		keycodes.DeleteKeycode(keycode)
	},
}

func init() {
	parseDeleteKeycodeFlag()
}

func parseDeleteKeycodeFlag() {
	DeleteKeycodeCmd.Flags().StringVar(&keycode, "keycode", "", "The keycode which will be to deleted in datahub.")
}
