package ui

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetPrompt(t *testing.T) {
	t.Run("with no hidden input", func(t *testing.T) {
		message := "Username: "
		prompt := GetPrompt(message, false)

		assert.Equal(t, message, prompt.Label)
	})

	t.Run("with hidden input", func(t *testing.T) {
		message := "Username: "
		mask := '*'
		prompt := GetPrompt(message, true)

		assert.Equal(t, message, prompt.Label)
		assert.Equal(t, mask, prompt.Mask)
	})
}

func TestGetSSHCommand(t *testing.T) {
	args := []string{"arg1", "arg2"}
	expectedArgs := []string{"ssh", "arg1", "arg2"}
	result := GetSSHCommand(args)

	assert.Equal(t, result.Args, expectedArgs)
	assert.Equal(t, result.Stdin, os.Stdin)
	assert.Equal(t, result.Stdout, os.Stdout)
}
