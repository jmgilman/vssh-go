package auth

//go:generate moq -out ../../internal/mocks/authinterface.go -pkg mocks . Auth
type Auth interface {
	GetData(map[string]*Detail) map[string]interface{}
	GetPath(map[string]*Detail) string
	AuthDetails() map[string]*Detail
}

type Detail struct {
	Value interface{}
	Prompt string
	Hidden bool
}