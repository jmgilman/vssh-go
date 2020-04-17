package auth

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewUserPassAuth(t *testing.T) {
	testUP := &UserPassAuth{
		username: "username",
		password: "password",
		mount: "userpass",
	}

	result := NewUserPassAuth("username", "password")
	assert.Equal(t, testUP, result)
}

func TestNewUserPassRadiusAuth(t *testing.T) {
	testUP := &UserPassAuth{
		username: "username",
		password: "password",
		mount: "radius",
	}

	result := NewUserPassRadiusAuth("username", "password")
	assert.Equal(t, testUP, result)
}

func TestUserPassAuth_GetPath(t *testing.T) {
	username := "username"
	mount := "mount"
	expected := fmt.Sprintf("auth/%s/login/%s", mount, username)
	testUP := &UserPassAuth{
		username: username,
		mount: mount,
	}

	result := testUP.GetPath()
	assert.Equal(t, expected, result)
}

func TestUserPassAuth_GetData(t *testing.T) {
	password := "test"
	testUP := &UserPassAuth{
		password: password,
	}

	result := testUP.GetData()
	assert.Equal(t, password, result["password"])
}
