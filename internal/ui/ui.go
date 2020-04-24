package ui

import (
	"github.com/jmgilman/vssh/auth"
	"github.com/manifoldco/promptui"
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