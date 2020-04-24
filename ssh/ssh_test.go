package ssh

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewSSHCommand(t *testing.T) {
	args := []string{"arg1", "arg2"}
	expectedArgs := []string{"ssh", "arg1", "arg2"}
	result := NewSSHCommand(args)

	assert.Equal(t, result.Args, expectedArgs)
	assert.Equal(t, result.Stdin, os.Stdin)
	assert.Equal(t, result.Stdout, os.Stdout)
}

func TestGetPublicKeyPath(t *testing.T) {
	path := "some/fake/path/key"

	t.Run("With identity file", func(t *testing.T) {
		pubKeyPath, err := GetPublicKeyPath(path)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "pub", pubKeyPath[len(pubKeyPath)-3:])
	})
	t.Run("Without identity file", func(t *testing.T) {
		pubKeyPath, err := GetPublicKeyPath("")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "pub", pubKeyPath[len(pubKeyPath)-3:])
	})
}

func TestGetPublicKeyCertPath(t *testing.T) {
	path := "/home/user/.ssh/id_rsa.pub"
	expected := "/home/user/.ssh/id_rsa-cert.pub"

	assert.Equal(t, expected, GetPublicKeyCertPath(path))
}
