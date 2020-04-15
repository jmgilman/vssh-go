package ui

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"syscall"
)

func GetInput(prompt string) (string, error) {
	fmt.Print(prompt + ": ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	result := scanner.Text()

	return result, nil
}

func GetInputHidden(prompt string) (string, error) {
	fmt.Print(prompt + ": ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))

	return string(bytePassword), err
}