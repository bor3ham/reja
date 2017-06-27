package instances

type Instance interface {
	GetID() string
	SetID(string)
	GetType() string
	SetValues(values []interface{})
	GetValues() []interface{}
}
