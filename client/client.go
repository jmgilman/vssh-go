// The client package contains the implementation of VaultClient which acts as a small wrapper around the Vault API
// client (github.com/hashicorp/vault/api).
package client

import (
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/jmgilman/vssh/auth"
)

// VaultClient is a small wrapper around the Vault API client. It provides additional functionality needed by vssh such
// as handling authentication a client and signing SSH public keys.
type VaultClient struct {
	api *api.Client
}

// NewClient returns a new VaultClient with the underlying API client configured with the given api.Config.
func NewClient(c *api.Config) (*VaultClient, error) {
	apiClient, err := api.NewClient(c)
	if err != nil {
		return &VaultClient{}, err
	}
	return &VaultClient{
		api: apiClient,
	}, nil
}

// NewClientWithAPI returns a new VaultClient with the underlying API client configured with the given api.Client.
func NewClientWithAPI(c *api.Client) *VaultClient {
	return &VaultClient{api: c}
}

// NewDefaultClient() returns a new VaultClient with the underlying API client configured with the Vault default values.
func NewDefaultClient() (*VaultClient, error) {
	return NewClient(api.DefaultConfig())
}

// Login takes an authentication type along with its associated details and attempts to authenticate against the
// configured Vault instance. If authentication is successful, the token returned from the Vault instance will be
// automatically set to the underlying API client.
func (c *VaultClient) Login(a auth.Auth, d map[string]*auth.Detail) error {
	secret, err := c.api.Logical().Write(a.GetPath(d), a.GetData(d))

	if err != nil {
		return err
	}

	if secret.Auth == nil {
		return fmt.Errorf("login returned an empty token")
	}

	c.api.SetToken(secret.Auth.ClientToken)
	return nil
}

// SignPubKey will use the underlying API client to attempt to sign the given SSH public key with the given role and
// mount point.
func (c *VaultClient) SignPubKey(mount string, role string, key []byte) (string, error) {
	var ssh *api.SSH
	// The SSH method sets the mount to its default value of "ssh"
	if mount == "" {
		ssh = c.api.SSH()
	} else {
		ssh = c.api.SSHWithMountPoint(mount)
	}

	data := map[string]interface{} {
		"public_key": string(key),
		"cert_type": "user",
	}

	// SignKey is a nice API wrapper which handles most of the logic for signing a key
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

// Authenticated performs a lookup of the underlying API client which by nature requires a valid token. If the lookup
// fails it will return false, indicating the client does not have a valid token. If the lookup succeeds, it returns
// true.
func (c *VaultClient) Authenticated() bool {
	_, err := c.api.Auth().Token().LookupSelf()
	if err != nil {
		return false
	} else {
		return true
	}
}

// Available checks if the configured Vault instance is either sealed or not initialized, returning false if either of
// those conditions are true.
func (c *VaultClient) Available() (bool, error) {
	status, err := c.api.Sys().SealStatus()
	if err != nil {
		return false, err
	}

	if !status.Sealed && status.Initialized {
		return true, nil
	}

	return false, nil
}

// SetConfigValues provides a method for setting the server and token of the underlying API client.
func (c *VaultClient) SetConfigValues(server string, token string) error {
	if server != "" {
		if err := c.api.SetAddress(server); err != nil {
			return err
		}
	}

	if token != "" {
		c.api.SetToken(token)
	}

	return nil
}

// Address returns the Vault instance address configured for the underlying API client.
func (c *VaultClient) Address() string {
	return c.api.Address()
}

// Token returns the token configured for the underlying API client.
func (c *VaultClient) Token() string {
	return c.api.Token()
}