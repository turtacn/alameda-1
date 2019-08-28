package keycodes

import (
	"fmt"
	Utils "github.com/containers-ai/alameda/datahub/tools/license-utils/pkg/utils"
	Keycodes "github.com/containers-ai/api/datahub/keycodes"
	"github.com/manifoldco/promptui"
)

var (
	datahubAddress *string
)

func KeycodeInit(address *string) {
	datahubAddress = address
}

func Executor() (string, error) {
	prompt := promptui.Select{
		Label: "Select Option",
		Items: []string{"Add", "Read", "Delete", "Activate", "Generate Registration Data", "Back"},
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Invalid input value %v\n", err)
		return "", err
	}

	switch result {
	case "Add":
		keycode, _ := Utils.InputText("Keycode")
		return "Add", AddKeycode(keycode)
	case "Read":
		keycode, _ := Utils.InputText("Keycode")
		return "Read", ListKeycodes(keycode)
	case "Delete":
		keycode, _ := Utils.InputText("Keycode")
		return "Delete", DeleteKeycode(keycode)
	case "Activate":
		filePath, _ := Utils.InputText("Registration File Path")
		return "Activate", Activate(filePath)
	case "Generate Registration Data":
		return "Generate Registration Data", GenerateRegistrationData()
	default:
		return "Back", nil
	}
}

func PrintKeycode(keycode *Keycodes.Keycode) {
	fmt.Println(fmt.Sprintf("Keycode: %s", keycode.Keycode))
	fmt.Println(fmt.Sprintf("KeycodeType: %s", keycode.KeycodeType))
	fmt.Println(fmt.Sprintf("KeycodeVersion: %d", keycode.KeycodeVersion))
	fmt.Println(fmt.Sprintf("ApplyTime: %s", keycode.ApplyTime))
	fmt.Println(fmt.Sprintf("ExpireTime: %s", keycode.ExpireTime))
	fmt.Println(fmt.Sprintf("Registered: %t", keycode.Registered))
	fmt.Println(fmt.Sprintf("LicenseState: %s", keycode.LicenseState))
	fmt.Println(fmt.Sprintf("Max Users: %d", keycode.Capacity.Users))
	fmt.Println(fmt.Sprintf("Max Hosts: %d", keycode.Capacity.Hosts))
	fmt.Println(fmt.Sprintf("Max Disks: %d", keycode.Capacity.Disks))
	fmt.Println(fmt.Sprintf("Diskprophet enabled: %t", keycode.Functionality.DiskProphet))
	fmt.Println(fmt.Sprintf("Workload enabled: %t", keycode.Functionality.Workload))
}
