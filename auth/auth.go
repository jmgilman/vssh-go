package auth

//go:generate moq -out ../internal/mocks/authinterface.go -pkg mocks . Auth
type Auth interface {
	Name() string
	GetData(map[string]*Detail) map[string]interface{}
	GetPath(map[string]*Detail) string
	AuthDetails() map[string]*Detail
}

type Detail struct {
	Value interface{}
	Prompt string
	Hidden bool
}

// Types is a map of every authentication type's name to its associated factory function
var Types = map[string]func() Auth{
	NewUserPassAuth().Name(): NewUserPassAuth,
	NewUserPassRadiusAuth().Name(): NewUserPassRadiusAuth,
}

func GetAuthNames() []string {
	names := make([]string, len(Types))

	i := 0
	for name := range Types {
		names[i] = name
		i++
	}

	return names
}