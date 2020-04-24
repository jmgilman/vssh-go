package ui_test

import (
	"github.com/jmgilman/vssh/auth"
	"github.com/jmgilman/vssh/internal/mocks"
	"github.com/jmgilman/vssh/internal/ui"
	"github.com/manifoldco/promptui"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAuthDetails2(t *testing.T) {
	expected := promptui.Select{
		Label: "test",
		Items: []string{"test", "test1"},
	}
	got := ui.NewSelectPrompt("test", []string{"test", "test1"})
	assert.Equal(t, expected.Label, got.Label)
	assert.Equal(t, expected.Items, got.Items)
}

func TestGetAuthDetails(t *testing.T) {
	var messages []string
	expectedMessages := []string{"Field1", "Field2"}
	prompter := func(message string, hidden bool) ui.Prompter {
		messages = append(messages, message)
		return &mocks.PrompterMock{
			RunFunc: func() (string, error) {
				return "test", nil
			},
		}
	}
	mockAuth := &mocks.AuthMock{
		AuthDetailsFunc: func() map[string]*auth.Detail {
			return map[string]*auth.Detail{
				"field1": &auth.Detail{
					Prompt: "Field1",
					Hidden: false,
				},
				"field2": &auth.Detail{
					Prompt: "Field2",
					Hidden: true,
				},
			}
		},
	}

	details, err := ui.GetAuthDetails(mockAuth, prompter)
	if err != nil {
		t.Fatal(err)
	}

	// Assert that the prompt was called with all given details
	for _, message := range messages {
		assert.Contains(t, expectedMessages, message)
	}

	// Assert that the return from Run() was put back into the details struct
	assert.Equal(t, "test", details["field1"].Value)
	assert.Equal(t, "test", details["field2"].Value)
}