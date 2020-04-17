package client

import (
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/jmgilman/vssh/pkg/auth"
)

//go:generate moq -out ../../internal/mocks/apiinterface.go -pkg mocks . API
type API interface {
	Logical() Writer
	SetToken(string)
	Token() string
}

//go:generate moq -out ../../internal/mocks/writerinterface.go -pkg mocks . Writer
type Writer interface {
	Write(string, map[string]interface{}) (*api.Secret, error)
}

type VaultClient struct {
	api API
}

func New(api API) *VaultClient {
	return &VaultClient{
		api: api,
	}
}

func (c *VaultClient) Login(a auth.Auth) error {
	secret, err := c.api.Logical().Write(a.GetPath(), a.GetData())

	if err != nil {
		return err
	}

	if secret.Auth == nil {
		return fmt.Errorf("login returned an empty token")
	}

	// Set the new client token
	c.api.SetToken(secret.Auth.ClientToken)
	return nil
}