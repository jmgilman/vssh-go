package auth

type Auth interface {
	GetData() (map[string]interface{}, error)
	GetPath() (string, error)
}