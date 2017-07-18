package schema

type Attribute interface {
	GetKey() string

	GetSelectDirectColumns() []string
	GetSelectDirectVariables() []interface{}

	ValidateFilters(map[string][]string) ([]Filter, error)
	GetFilterWhere(int, map[string][]string) ([]string, []interface{})

	DefaultFallback(interface{}, interface{}) (interface{}, error)
	Validate(interface{}) (interface{}, error)

	GetInsertColumns(interface{}) []string
	GetInsertValues(interface{}) []interface{}
}
