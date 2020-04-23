package auth

import (
	"fmt"
)

type UserPassAuth struct {
	name string
	mount string
}

func NewUserPassAuth() Auth {
	return &UserPassAuth{
		name: "Userpass",
		mount: "userpass",
	}
}

func NewUserPassRadiusAuth() Auth {
	return &UserPassAuth{
		name: "Radius",
		mount: "radius",
	}
}

func (u *UserPassAuth) Name() string {
	return u.name
}

func (u *UserPassAuth) AuthDetails() map[string]*Detail {
	return map[string]*Detail {
		"username": {
			Prompt: "Username: ",
			Hidden: false,
		},
		"password": {
			Prompt: "Password: ",
			Hidden: true,
		},
	}
}

func (u *UserPassAuth) GetPath(details map[string]*Detail) string {
	return fmt.Sprintf("auth/%s/login/%s", u.mount, details["username"].Value)
}

func (u *UserPassAuth) GetData(details map[string]*Detail) map[string]interface{} {
	return map[string]interface{}{
		"password": details["password"].Value,
	}
}