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

func NewSSHCommand(args []string) *exec.Cmd {
	c := exec.Command("ssh", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c
}

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

func IsCertificateValid(certPath string) (bool, error) {
	if _, err := os.Stat(certPath); !os.IsNotExist(err) {
		signedKeyBytes, err := ioutil.ReadFile(certPath)
		if err != nil {
			return false, err
		}

		cert, _, _, _, err := cssh.ParseAuthorizedKey(signedKeyBytes)
		if err != nil {
			return false, err
		}

		validBefore := int64(cert.(*cssh.Certificate).ValidBefore)
		validAfter := int64(cert.(*cssh.Certificate).ValidAfter)
		now := time.Now().Unix()
		if now < validBefore && now > validAfter {
			return true, nil
		}
	}

	return false, nil
}