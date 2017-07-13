package schema

type Attribute interface {
	GetKey() string

	GetSelectDirectColumns() []string
	GetSelectDirectVariables() []interface{}

	ValidateFilters(map[string][]string) (map[string][]string, error)
	GetFilterWhere(int, map[string][]string) ([]string, []interface{})
	GetFilterAnnotate(map[string]string) string

	DefaultFallback(interface{}, interface{}) (interface{}, error)
	Validate(interface{}) (interface{}, error)

	GetInsertColumns(interface{}) []string
	GetInsertValues(interface{}) []interface{}
}
