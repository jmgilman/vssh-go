package ui

import (
	"github.com/manifoldco/promptui"
	"os"
	"os/exec"
)

func GetPrompt(message string, hidden bool) *promptui.Prompt {
	if !hidden {
		return &promptui.Prompt{
			Label: message,
		}
	} else {
		return &promptui.Prompt{
			Label: message,
			Mask: '*',
		}
	}
}

func GetSSHCommand(args []string) *exec.Cmd {
	c := exec.Command("ssh", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c
}