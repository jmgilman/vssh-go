// The ui package contains functions for retrieving information from an end-user via the CLI.
package ui

import (
	"github.com/jmgilman/vssh/auth"
	"github.com/manifoldco/promptui"
)

//go:generate moq -out ../../internal/mocks/prompterinterface.go -pkg mocks . Prompter
// Prompter is used for testing purposes.
type Prompter interface {
	Run() (string, error)
}

// NewPrompt returns a promptui.Prompt which has its prompt message configured to the given message and adds an
// additional character mask if hidden is set to true.
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

// NewSelectPrompt returns a promptui.SelectPrompt with its prompt message configured to the given message and the
// available options for the user to select configured to the given string slice.
func NewSelectPrompt(message string, options []string) *promptui.Select {
	return &promptui.Select{
		Label: message,
		Items: options,
	}
}

// GetAuthDetails retrieves the authentication details from the given authentication type and proceeds to prompt the
// user to provide input for each of the retrieved details. It returns the detail map configured with the input data
// from the end-user.
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