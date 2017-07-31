package schema

type Attribute interface {
	GetKey() string

	GetSelectDirectColumns() []string
	GetSelectDirectVariables() []interface{}

	GetOrderMap() map[string]string

	AvailableFilters() []interface{}
	ValidateFilters(map[string][]string) ([]Filter, error)
	GetFilterWhere(int, map[string][]string) ([]string, []interface{})

	DefaultFallback(interface{}, interface{}) (interface{}, error)
	Validate(interface{}) (interface{}, error)
	ValidateUpdate(interface{}, interface{}) (interface{}, error)

	GetInsert(interface{}) ([]string, []interface{})
}
