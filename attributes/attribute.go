package attributes

type Attribute interface {
	GetSelectDirectColumns() []string
	GetSelectDirectVariables() []interface{}

	ValidateNew(interface{}) (interface{}, error)
}
