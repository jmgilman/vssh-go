package client_test

import (
	"encoding/base64"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/builtin/credential/userpass"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/vault"
	"github.com/jmgilman/vssh/client"
	"github.com/jmgilman/vssh/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net"
	"os"
	"testing"
)

type ClientTestSuite struct {
	suite.Suite
	apiClient *api.Client
	handler net.Listener
	rootToken string
	keys [][]byte
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (suite *ClientTestSuite) SetupTest() {
	// Initialize an in-memory Vault server
	suite.handler, suite.apiClient, suite.keys = suite.NewVaultServer()
	suite.rootToken = suite.apiClient.Token()
}

func (suite *ClientTestSuite) TearDownTest() {
	// Cleanup HTTP handler
	if err := suite.handler.Close(); err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *ClientTestSuite) NewVaultServer() (net.Listener, *api.Client, [][]byte) {
	t := suite.T()
	t.Helper()

	// Create an in-memory, unsealed core with userpass auth plugin enabled
	coreConfig := &vault.CoreConfig{
		CredentialBackends: map[string]logical.Factory{
			"userpass": userpass.Factory,
		},
	}
	core, keyShares, rootToken := vault.TestCoreUnsealedWithConfig(t, coreConfig)

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
	_, err = apiClient.Logical().Write("auth/userpass/users/test", suite.NewCreds("password"))
	if err != nil {
		t.Fatal(err)
	}

	return ln, apiClient, keyShares
}

func (suite *ClientTestSuite) NewCreds(password string) map[string]interface{} {
	suite.T().Helper()
	return map[string]interface{} {
		"password": password,
	}
}

func (suite *ClientTestSuite) NewMockAuth(password string) *mocks.AuthMock {
	suite.T().Helper()
	return &mocks.AuthMock{
		GetPathFunc: func() string {return "auth/userpass/login/test"},
		GetDataFunc: func() map[string]interface{} {return suite.NewCreds(password)},
	}
}

func (suite *ClientTestSuite) TestNewClient() {
	config := &api.Config{
		Address: "http://127.1.1:8200",
	}
	vaultClient, err := client.NewClient(config)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), vaultClient.Address(), config.Address)
}

func (suite *ClientTestSuite) TestNewDefaultClient() {
	// The Vault default config pulls address from the VAULT_ADDR environment variable
	if err := os.Setenv("VAULT_ADDR", "http://127.1.1:8200"); err != nil {
		suite.T().Fatal(err)
	}
	vaultClient, err := client.NewDefaultClient()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), vaultClient.Address(), "http://127.1.1:8200")
}

func (suite *ClientTestSuite) TestVaultClient_Login() {
	// Setup helper objects
	vaultClient := client.NewClientWithAPI(suite.apiClient)
	t := suite.T()

	t.Run("Test with valid login", func(t *testing.T) {
		suite.apiClient.SetToken("")

		err := vaultClient.Login(suite.NewMockAuth("password"))
		assert.Nil(t, err)
		assert.NotEmpty(t, vaultClient.Token())
	})

	t.Run("Test with invalid login", func(t *testing.T) {
		suite.apiClient.SetToken("")

		err := vaultClient.Login(suite.NewMockAuth("wrongpassword"))
		assert.Empty(t, vaultClient.Token())
		switch err := err.(type) {
		case *api.ResponseError:
			assert.Equal(t, err.StatusCode, 400)
		default:
			t.Fatal(err)
		}
	})
}

func (suite *ClientTestSuite) TestAuthenticated() {
	t := suite.T()
	vaultClient := client.NewClientWithAPI(suite.apiClient)

	t.Run("Test with valid credentials", func(t *testing.T) {
		suite.apiClient.SetToken(suite.rootToken)
		assert.True(t, vaultClient.Authenticated())
	})
	t.Run("Test with invalid credentials", func(t *testing.T) {
		suite.apiClient.SetToken("")
		assert.False(t, vaultClient.Authenticated())
	})
}

func (suite *ClientTestSuite) TestAvailable() {
	t := suite.T()
	vaultClient := client.NewClientWithAPI(suite.apiClient)
	t.Run("Test with an available vault", func(t *testing.T) {
		status, err := vaultClient.Available()
		assert.Nil(t, err)
		assert.True(t, status)
	})
	t.Run("Test with a sealed vault", func(t *testing.T) {
		// Seal the vault
		if err := suite.apiClient.Sys().Seal(); err != nil {
			suite.T().Fatal(err)
		}
		status, err := vaultClient.Available()
		assert.Nil(t, err)
		assert.False(t, status)

		// Unseal vault
		for _, key := range suite.keys {
			encodedKey := base64.StdEncoding.EncodeToString(key)
			status, err := suite.apiClient.Sys().Unseal(encodedKey)
			if err != nil {
				t.Fatal(err)
			}
			if !status.Sealed {
				break
			}
		}
	})

	// TODO(jmgilman): Implement a test for an uninitialized vault
}