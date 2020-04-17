package client_test

import (
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/jmgilman/vssh/internal/mocks"
	"github.com/jmgilman/vssh/pkg/client"
	"github.com/stretchr/testify/assert"
	"testing"
)

func NewMockAPI(write *mocks.WriterMock) *mocks.APIMock {
	mockAPI := mocks.APIMock{}
	mockAPI.LogicalFunc = func() client.Writer {
		return write
	}
	mockAPI.SetTokenFunc = func(in1 string) {}
	return &mockAPI
}

func NewMockWrite(token string, err error) *mocks.WriterMock {
	return &mocks.WriterMock{WriteFunc: func(in1 string, in2 map[string]interface{}) (*api.Secret, error) {
		return &api.Secret{
			Auth: &api.SecretAuth{
				ClientToken: token,
			},
		}, err
	}}
}

func NewMockWriteEmpty(token string, err error) *mocks.WriterMock {
	return &mocks.WriterMock{WriteFunc: func(in1 string, in2 map[string]interface{}) (*api.Secret, error) {
		return &api.Secret{}, err
	}}
}

func NewMockAuth() *mocks.AuthMock {
	return &mocks.AuthMock{
		GetDataFunc: func() map[string]interface{} { return make(map[string]interface{})},
		GetPathFunc: func() string { return ""},
	}
}

func TestVaultClient_Login(t *testing.T) {
	t.Run("test with no error", func(t *testing.T) {
		// Generate mocks
		mockWrite := NewMockWrite("test123", nil)
		mockAPI := NewMockAPI(mockWrite)
		mockAuth := NewMockAuth()

		// Setup new VaultClient using mock
		vaultClient := client.New(mockAPI)

		// Test a successful login
		err := vaultClient.Login(mockAuth)
		assert.Nil(t, err)
		assert.NotEmpty(t, mockAuth.GetDataCalls())
		assert.NotEmpty(t, mockAuth.GetPathCalls())
		assert.Equal(t, mockAPI.SetTokenCalls()[0].In1, "test123")
	})
	t.Run("test with error", func(t *testing.T) {
		// Generate mocks
		fakeErr := fmt.Errorf("error")
		mockWrite := NewMockWrite("test123", fakeErr)
		mockAPI := NewMockAPI(mockWrite)
		mockAuth := NewMockAuth()

		// Setup new VaultClient using mocks
		vaultClient := client.New(mockAPI)

		// Test an api error
		err := vaultClient.Login(mockAuth)
		assert.Equal(t, err, fakeErr)
		assert.NotEmpty(t, mockAuth.GetDataCalls())
		assert.NotEmpty(t, mockAuth.GetPathCalls())
		assert.Empty(t, mockAPI.SetTokenCalls())
	})
	t.Run("test with empty response", func(t *testing.T) {
		// Generate mocks
		mockWrite := NewMockWriteEmpty("test123", nil)
		mockAPI := NewMockAPI(mockWrite)
		mockAuth := NewMockAuth()

		// Setup new VaultClient using mocks
		vaultClient := client.New(mockAPI)

		// Test an empty response
		err := vaultClient.Login(mockAuth)
		assert.NotEmpty(t, err)
		assert.NotEmpty(t, mockAuth.GetDataCalls())
		assert.NotEmpty(t, mockAuth.GetPathCalls())
		assert.Empty(t, mockAPI.SetTokenCalls())
	})
}