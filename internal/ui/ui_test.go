package ui

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewPrompt(t *testing.T) {
	t.Run("with no hidden input", func(t *testing.T) {
		message := "Username: "
		prompt := NewPrompt(message, false)

		assert.Equal(t, message, prompt.Label)
	})

	t.Run("with hidden input", func(t *testing.T) {
		message := "Username: "
		mask := '*'
		prompt := NewPrompt(message, true)

		assert.Equal(t, message, prompt.Label)
		assert.Equal(t, mask, prompt.Mask)
	})
}

func TestNewSSHCommand(t *testing.T) {
	args := []string{"arg1", "arg2"}
	expectedArgs := []string{"ssh", "arg1", "arg2"}
	result := NewSSHCommand(args)

	assert.Equal(t, result.Args, expectedArgs)
	assert.Equal(t, result.Stdin, os.Stdin)
	assert.Equal(t, result.Stdout, os.Stdout)
}
