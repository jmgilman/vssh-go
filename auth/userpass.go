package auth

import "fmt"

type UserPassAuth struct {
	username string
	password string
	mount string
}

func NewUserPassAuth(username string, password string) *UserPassAuth {
	return &UserPassAuth{
		username,
		password,
		"userpass",
	}
}

func NewUserPassRadiusAuth(username string, password string) *UserPassAuth {
	return &UserPassAuth{
		username: username,
		password: password,
		mount: "radius",
	}
}

func (u *UserPassAuth) GetPath() string {
	return fmt.Sprintf("auth/%s/login/%s", u.mount, u.username)
}

func (u *UserPassAuth) GetData() map[string]interface{} {
	return map[string]interface{}{
		"password": u.password,
	}
}