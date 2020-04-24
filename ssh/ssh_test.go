package ssh

import (
	"github.com/stretchr/testify/assert"
	cssh "golang.org/x/crypto/ssh"
	"os"
	"testing"
	"time"
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

func TestIsCertificateValid(t *testing.T) {
	t.Run("With valid time", func(t *testing.T) {
		cert := &cssh.Certificate{
			ValidBefore: uint64(time.Now().Unix()) + 1000,
			ValidAfter: uint64(time.Now().Unix()) - 1000,
		}
		assert.True(t, IsCertificateValid(cert))
	})
	t.Run("With invalid time", func(t *testing.T) {
		cert := &cssh.Certificate{
			ValidBefore: uint64(time.Now().Unix()) - 1000,
			ValidAfter: uint64(time.Now().Unix()) + 1000,
		}
		assert.False(t, IsCertificateValid(cert))
	})
}