package client

import (
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/jmgilman/vssh/auth"
)

type VaultClient struct {
	api *api.Client
}

func NewClient(c *api.Config) (*VaultClient, error) {
	apiClient, err := api.NewClient(c)
	if err != nil {
		return &VaultClient{}, err
	}
	return &VaultClient{
		api: apiClient,
	}, nil
}

func NewClientWithAPI(c *api.Client) *VaultClient {
	return &VaultClient{api: c}
}

func NewDefaultClient() (*VaultClient, error) {
	return NewClient(api.DefaultConfig())
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

func (c *VaultClient) Address() string {
	return c.api.Address()
}

func (c *VaultClient) Token() string {
	return c.api.Token()
}