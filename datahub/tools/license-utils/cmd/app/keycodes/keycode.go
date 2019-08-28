package keycodes

import (
	"github.com/spf13/cobra"
)

var KeycodeCmd = &cobra.Command{
	Use:   "keycode",
	Short: "datahub keycode utilities",
	Long:  "",
}

var (
	keycode  string
	filePath string
)

func init() {
	KeycodeCmd.AddCommand(AddKeycodeCmd)
	KeycodeCmd.AddCommand(ReadKeycodeCmd)
	KeycodeCmd.AddCommand(DeleteKeycodeCmd)
	KeycodeCmd.AddCommand(GenerateRegistrationDataCmd)
	KeycodeCmd.AddCommand(ActivateCmd)
}
