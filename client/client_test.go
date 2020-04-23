package client_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/builtin/credential/userpass"
	"github.com/hashicorp/vault/builtin/logical/ssh"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/vault"
	"github.com/jmgilman/vssh/auth"
	"github.com/jmgilman/vssh/client"
	"github.com/jmgilman/vssh/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	cssh "golang.org/x/crypto/ssh"
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

func (suite *ClientTestSuite) SetupSuite() {
	// Initialize an in-memory Vault server
	suite.handler, suite.apiClient, suite.keys = suite.NewVaultServer()
	suite.rootToken = suite.apiClient.Token()
}

func (suite *ClientTestSuite) TearDownSuite() {
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
		LogicalBackends: map[string]logical.Factory {
			"ssh": ssh.Factory,
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

	// Setup SSH backend
	roleData := map[string]interface{} {
		"allow_user_certificates": true,
		"allowed_users": "*",
		"key_type": "ca",
		"ttl": "30m0s",
	}
	err = apiClient.Sys().Mount("ssh", &api.MountInput{Type: "ssh"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = apiClient.Logical().Write("ssh/config/ca", map[string]interface{}{ "generate_signing_key": true})
	if err != nil {
		t.Fatal(err)
	}
	_, err = apiClient.Logical().Write("ssh/roles/test", roleData)
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
		GetPathFunc: func(map[string]*auth.Detail) string {return "auth/userpass/login/test"},
		GetDataFunc: func(map[string]*auth.Detail) map[string]interface{} {return suite.NewCreds(password)},
	}
}

func (suite *ClientTestSuite) EncodeSSHPrivateKey(key *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)
}

func (suite *ClientTestSuite) NewSSHPubKey() ([]byte, error) {
	suite.T().Helper()
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return []byte{}, err
	}
	publicKey, _ := cssh.NewPublicKey(&privateKey.PublicKey)
	return cssh.MarshalAuthorizedKey(publicKey), nil
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
	details := map[string]*auth.Detail{} // Can use empty details since we already override the functions

	t.Run("Test with valid login", func(t *testing.T) {
		suite.apiClient.SetToken("")

		err := vaultClient.Login(suite.NewMockAuth("password"), details)
		assert.Nil(t, err)
		assert.NotEmpty(t, vaultClient.Token())
	})

	t.Run("Test with invalid login", func(t *testing.T) {
		suite.apiClient.SetToken("")

		err := vaultClient.Login(suite.NewMockAuth("wrongpassword"), details)
		assert.Empty(t, vaultClient.Token())
		switch err := err.(type) {
		case *api.ResponseError:
			assert.Equal(t, err.StatusCode, 400)
		default:
			t.Fatal(err)
		}
	})
}

func (suite *ClientTestSuite) TestSignPubKey() {
	suite.apiClient.SetToken(suite.rootToken)
	vaultClient := client.NewClientWithAPI(suite.apiClient)

	pubKey, err := suite.NewSSHPubKey()
	if err != nil {
		suite.T().Fatal(err)
	}

	result, err := vaultClient.SignPubKey("ssh", "test", pubKey)
	assert.Nil(suite.T(), err)
	assert.NotEmpty(suite.T(), result)
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
		suite.apiClient.SetToken(suite.rootToken)
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