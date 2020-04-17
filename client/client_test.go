package client_test

import (
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/builtin/credential/userpass"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/vault"
	"github.com/jmgilman/vssh/client"
	"github.com/jmgilman/vssh/internal/mocks"
	"github.com/stretchr/testify/assert"
	"net"
	"os"
	"testing"
)

func NewVaultServer(t *testing.T) (net.Listener, *api.Client) {
	t.Helper()

	// Create an in-memory, unsealed core with userpass auth plugin enabled
	coreConfig := &vault.CoreConfig{
		CredentialBackends: map[string]logical.Factory{
			"userpass": userpass.Factory,
		},
	}
	core, keyShares, rootToken := vault.TestCoreUnsealedWithConfig(t, coreConfig)
	_ = keyShares

	// Start an HTTP server for the core.
	ln, addr := http.TestServer(t, core)

	// Create a client that talks to the server, initially authenticating with
	// the root token.
	conf := api.DefaultConfig()
	conf.Address = addr

	apiClient, err := api.NewClient(conf)
	if err != nil {
		t.Fatal(err)
	}

	// Setup test user account
	apiClient.SetToken(rootToken)
	err = apiClient.Sys().EnableAuthWithOptions("userpass", &api.EnableAuthOptions{Type: "userpass"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = apiClient.Logical().Write("auth/userpass/users/test", NewCreds(t, "password"))
	if err != nil {
		t.Fatal(err)
	}

	return ln, apiClient
}

func NewCreds(t *testing.T, password string) map[string]interface{} {
	t.Helper()
	return map[string]interface{} {
		"password": password,
	}
}

func NewMockAuth(t *testing.T, password string) *mocks.AuthMock {
	t.Helper()
	return &mocks.AuthMock{
		GetPathFunc: func() string {return "auth/userpass/login/test"},
		GetDataFunc: func() map[string]interface{} {return NewCreds(t, password)},
	}
}

func TestNewClient(t *testing.T) {
	config := &api.Config{
		Address: "http://127.1.1:8200",
	}
	vaultClient, err := client.NewClient(config)
	assert.Nil(t, err)
	assert.Equal(t, vaultClient.Address(), config.Address)
}

func TestNewDefaultClient(t *testing.T) {
	// The Vault default config pulls address from the VAULT_ADDR environment variable
	if err := os.Setenv("VAULT_ADDR", "http://127.1.1:8200"); err != nil {
		t.Fatal(err)
	}
	vaultClient, err := client.NewDefaultClient()
	assert.Nil(t, err)
	assert.Equal(t, vaultClient.Address(), "http://127.1.1:8200")
}

func TestVaultClient_Login(t *testing.T) {
	// Setup helper objects
	ln, apiClient := NewVaultServer(t)
	vaultClient := client.NewClientWithAPI(apiClient)

	t.Run("Test with valid login", func(t *testing.T) {
		apiClient.SetToken("")

		err := vaultClient.Login(NewMockAuth(t, "password"))
		assert.Nil(t, err)
		assert.NotEmpty(t, vaultClient.Token())
	})

	t.Run("Test with invalid login", func(t *testing.T) {
		apiClient.SetToken("")

		err := vaultClient.Login(NewMockAuth(t, "wrongpassword"))
		assert.Empty(t, vaultClient.Token())
		switch err := err.(type) {
		case *api.ResponseError:
			assert.Equal(t, err.StatusCode, 400)
		default:
			t.Fatal(err)
		}
	})

	// Cleanup
	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}
}