package auth

//go:generate moq -out ../../internal/mocks/authinterface.go -pkg mocks . Auth
type Auth interface {
	GetData() map[string]interface{}
	GetPath() string
}