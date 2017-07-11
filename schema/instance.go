package schema

type Instance interface {
	GetID() string
	SetID(string)
	GetType() string
	SetValues(values map[string]interface{})
	GetValues() map[string]interface{}
}
