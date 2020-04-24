package ui

import (
	"fmt"
	"github.com/jmgilman/vssh/auth"
	"github.com/manifoldco/promptui"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:generate moq -out ../../internal/mocks/prompterinterface.go -pkg mocks . Prompter
type Prompter interface {
	Run() (string, error)
}

func NewPrompt(message string, hidden bool) Prompter {
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

func NewSelectPrompt(message string, options []string) *promptui.Select {
	return &promptui.Select{
		Label: message,
		Items: options,
	}
}

func NewSSHCommand(args []string) *exec.Cmd {
	c := exec.Command("ssh", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c
}

func GetAuthDetails(a auth.Auth, prompterFactory func(message string, hidden bool) Prompter) (map[string]*auth.Detail, error) {
	details := a.AuthDetails()
	for _, detail := range details {
		prompt := prompterFactory(detail.Prompt, detail.Hidden)
		result, err := prompt.Run()
		if err != nil {
			return map[string]*auth.Detail{}, err
		}

		detail.Value = result
	}

	return details, nil
}

func GetPublicKeyPath(identity string) (publicKeyPath string, err error) {
	if identity == "" {
		defaultPubKeyPath, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Failed to get user home directory: ", err)
			os.Exit(1)
		}
		publicKeyPath = filepath.Join(defaultPubKeyPath, ".ssh/id_rsa.pub")
	} else {
		publicKeyPath = identity + ".pub"
	}
	return
}

func GetPublicKeyCertPath(pubKeyPath string) string {
	baseName := filepath.Base(pubKeyPath)
	baseExt := filepath.Ext(baseName)
	newName := strings.Split(baseName, baseExt)[0] + "-cert" + baseExt
	return filepath.Join(filepath.Dir(pubKeyPath), newName)
}