package ui

import (
	"github.com/manifoldco/promptui"
	"os"
	"os/exec"
)

func GetInput(message string) (result string, err error) {
	prompt := promptui.Prompt{
		Label: message,
	}
	result, err = prompt.Run()
	return
}

func GetInputHidden(message string) (result string, err error) {
	prompt := promptui.Prompt{
		Label:    "message",
		Mask:     '*',
	}

	result, err = prompt.Run()
	return
}

func CallSSH(args []string) {
	c := exec.Command("ssh", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Run()
}