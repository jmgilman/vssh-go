// auth contains functions and abstractions for authenticating against a Vault instance.
// It is built to be extendable by providing an interface that is accepted by the client Login() function for easily
// adding additional forms of authentication not currently supported by the package.
package auth

//go:generate moq -out ../internal/mocks/authinterface.go -pkg mocks . Auth
// Auth represents a form of authenticating with a Vault instance. See UserPassAuth for an example of how to properly
// implement this interface.
type Auth interface {
	Name() string
	GetData(map[string]*Detail) map[string]interface{}
	GetPath(map[string]*Detail) string
	AuthDetails() map[string]*Detail
}

// Detail represents a piece of information given by the end-user and required for performing authentication.
type Detail struct {
	Value interface{}
	Prompt string
	Hidden bool
}

// Types is a map of every authentication type's name to its associated factory function.
var Types = map[string]func() Auth{
	NewUserPassAuth().Name(): NewUserPassAuth,
	NewUserPassRadiusAuth().Name(): NewUserPassRadiusAuth,
}

// GetAuthNames returns the name of every type of authentication currently supported by the auth package.
func GetAuthNames() []string {
	names := make([]string, len(Types))

	i := 0
	for name := range Types {
		names[i] = name
		i++
	}

	return names
}