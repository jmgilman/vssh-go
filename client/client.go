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

func (c *VaultClient) Login(a auth.Auth, d map[string]*auth.Detail) error {
	secret, err := c.api.Logical().Write(a.GetPath(d), a.GetData(d))

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

func (c *VaultClient) SignPubKey(mount string, role string, key []byte) (string, error) {
	var ssh *api.SSH
	if mount == "" {
		ssh = c.api.SSH()
	} else {
		ssh = c.api.SSHWithMountPoint(mount)
	}

	data := map[string]interface{} {
		"public_key": string(key),
		"cert_type": "user",
	}

	result, err := ssh.SignKey(role, data)
	if err != nil {
		return "", err
	}

	if result == nil || result.Data == nil {
		return "", fmt.Errorf("no key was returned from the server")
	}

	signedKey, ok := result.Data["signed_key"].(string)
	if !ok || signedKey == "" {
		return "", fmt.Errorf("no key was returned from the server")
	}

	return signedKey, nil
}

func (c *VaultClient) Authenticated() bool {
	_, err := c.api.Auth().Token().LookupSelf()
	if err != nil {
		return false
	} else {
		return true
	}
}

func (c *VaultClient) Available() (bool, error) {
	status, err := c.api.Sys().SealStatus()
	if err != nil {
		return false, err
	}

	// Verify the Vault is not sealed and has been initialized
	if !status.Sealed && status.Initialized {
		return true, nil
	}

	return false, nil
}

func (c *VaultClient) SetConfigValues(server string, token string) {
	if server != "" {
		c.api.SetAddress(server)
	}

	if token != "" {
		c.api.SetToken(token)
	}
}

func (c *VaultClient) Address() string {
	return c.api.Address()
}

func (c *VaultClient) Token() string {
	return c.api.Token()
}