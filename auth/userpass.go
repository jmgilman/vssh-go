package auth

import (
	"fmt"
)

// UserPassAuth represents a form of authentication that takes a username and password.
type UserPassAuth struct {
	name string
	mount string
}

// NewUserPassAuth returns a new UserPassAuth struct with the name and mount already configured.
func NewUserPassAuth() Auth {
	return &UserPassAuth{
		name: "Userpass",
		mount: "userpass",
	}
}

// NewUserPassRadiusAuth returns a new UserPassAuth struct with the name and mount already configured for Radius.
func NewUserPassRadiusAuth() Auth {
	return &UserPassAuth{
		name: "Radius",
		mount: "radius",
	}
}

// Name returns the name of the authentication type. This is used when building a list of supported authentication
// types and should be a user friendly name.
func (u *UserPassAuth) Name() string {
	return u.name
}

// AuthDetails returns a map of detail names to their respective auth.Detail struct. This is used by the ui package to
// automatically collect the necessary authentication details required for this authentication type from the end-user.
// For example, the UserPassAuth type asks for the username and password for logging in.
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

// GetPath returns the Vault path to write to for performing this type of authentication
// (i.e. auth/userpass/login/user).
func (u *UserPassAuth) GetPath(details map[string]*Detail) string {
	return fmt.Sprintf("auth/%s/login/%s", u.mount, details["username"].Value)
}

// GetData returns a map of JSON data that will be written to the path returned by GetPath.
func (u *UserPassAuth) GetData(details map[string]*Detail) map[string]interface{} {
	return map[string]interface{}{
		"password": details["password"].Value,
	}
}