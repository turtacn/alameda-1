package setup

import (
	"fmt"
	"github.com/manifoldco/promptui"
)

func InputAddress() (string, error) {
	label := fmt.Sprintf("Datahub address(%s)", *datahubAddress)

	prompt := promptui.Prompt{
		Label:    label,
		Validate: validate,
	}
	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "", err
	}

	if result == "" {
		address := *datahubAddress
		return address, nil
	}

	return result, nil
}

func validate(input string) error {
	return nil
}
