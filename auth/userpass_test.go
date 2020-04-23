package auth

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewUserPassAuth(t *testing.T) {
	testUP := &UserPassAuth{
		mount: "userpass",
	}

	result := NewUserPassAuth()
	assert.Equal(t, testUP, result)
}

func TestNewUserPassRadiusAuth(t *testing.T) {
	testUP := &UserPassAuth{
		mount: "radius",
	}

	result := NewUserPassRadiusAuth()
	assert.Equal(t, testUP, result)
}

func TestUserPassAuth_GetPath(t *testing.T) {
	username := "username"
	mount := "userpass"
	expected := fmt.Sprintf("auth/%s/login/%s", mount, username)
	testUP := NewUserPassAuth()
	details := testUP.AuthDetails()
	details["username"].Value = username

	result := testUP.GetPath(details)
	assert.Equal(t, expected, result)
}

func TestUserPassAuth_GetData(t *testing.T) {
	password := "password"
	testUP := NewUserPassAuth()
	details := testUP.AuthDetails()
	details["password"].Value = password

	result := testUP.GetData(details)
	assert.Equal(t, password, result["password"])
}
