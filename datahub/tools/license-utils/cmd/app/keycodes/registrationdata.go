package keycodes

import (
	"github.com/containers-ai/alameda/datahub/tools/license-utils/pkg/keycodes"
	"github.com/spf13/cobra"
)

var GenerateRegistrationDataCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate registration data",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		keycodes.GenerateRegistrationData()
	},
}

func init() {
}
