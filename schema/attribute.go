package schema

type Attribute interface {
	GetSelectDirectColumns() []string
	GetSelectDirectVariables() []interface{}

	DefaultFallback(interface{}, interface{}) (interface{}, error)
	Validate(interface{}) (interface{}, error)

	GetInsertColumns(interface{}) []string
	GetInsertValues(interface{}) []interface{}
}
