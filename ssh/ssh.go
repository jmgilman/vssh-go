package ssh

import (
	"fmt"
	cssh "golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// NewSSHCommand returns a exec.Cmd type preconfigured to run the ssh binary using the given args and with all standard
// inputs/outputs configured to redirect the process to the end-user.
func NewSSHCommand(args []string) *exec.Cmd {
	c := exec.Command("ssh", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c
}

// GetPublicKey takes a path to a private key and finds its associated public key, reading it into memory and returning
// its content in byte form.
func GetPublicKey(identity string) (string, []byte, error) {
	publicKeyPath, err := GetPublicKeyPath(identity)
	if err != nil {
		return "", []byte{}, err
	}

	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		return "", []byte{}, err
	}

	data, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return "", []byte{}, err
	}

	return publicKeyPath, data, nil
}

// GetCertificate parses the SSH certificate at certPath and returns it as a ssh.Certificate.
func GetCertificate(certPath string) (*cssh.Certificate, error) {
	signedKeyBytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		return &cssh.Certificate{}, err
	}

	cert, _, _, _, err := cssh.ParseAuthorizedKey(signedKeyBytes)
	if err != nil {
		return &cssh.Certificate{}, err
	}

	return cert.(*cssh.Certificate), nil
}

// GetPublicKeyPath takes the path to a private key and returns the path to its associated public key. If the given
// path is empty, it defaults to returning the public key for $HOME/.ssh/id_rsa.
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

// GetPublicKeyCertPath takes the path to a SSH public key and returns the path to the associated signed certificate.
// For example, given $HOME/.ssh/id_rsa.pub it would return $HOME/.ssh/id_rsa-cert.pub.
func GetPublicKeyCertPath(pubKeyPath string) string {
	baseName := filepath.Base(pubKeyPath)
	baseExt := filepath.Ext(baseName)
	newName := strings.Split(baseName, baseExt)[0] + "-cert" + baseExt
	return filepath.Join(filepath.Dir(pubKeyPath), newName)
}

// IsCertificateValid takes a SSH certificate and returns whether or not it is expired (TTL has been exceeded).
func IsCertificateValid(cert *cssh.Certificate) bool {
	validBefore := int64(cert.ValidBefore)
	validAfter := int64(cert.ValidAfter)
	now := time.Now().Unix()
	if now < validBefore && now > validAfter {
		return true
	}

	return false
}