package utils

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"os"
	"os/exec"
)

func ClearScreen(showPrompt bool) {
	if showPrompt {
		contPrompt := promptui.Prompt{
			Label: "Press Enter to continue",
		}
		contPrompt.Run()
	}

	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func InputText(text string) (string, error) {
	prompt := promptui.Prompt{
		Label:    text,
		Validate: validate,
	}
	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "", err
	}

	return result, nil
}

func validate(input string) error {
	return nil
}
