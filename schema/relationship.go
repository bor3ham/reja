package schema

type Relationship interface {
	GetKey() string
	GetType() string

	GetSelectDirectColumns() []string
	GetSelectDirectVariables() []interface{}
	GetSelectExtraColumns() []string
	GetSelectExtraVariables() []interface{}

	AvailableFilters() []string
	ValidateFilters(map[string][]string) ([]Filter, error)
	GetFilterWhere(int, map[string][]string) ([]string, []interface{})

	GetDefaultValue() interface{}
	GetValues(
		Context,
		*Model,
		[]string,
		[][]interface{},
	) (
		map[string]interface{},
		map[string]map[string][]string,
	)

	GetInsertColumns(interface{}) []string
	GetInsertValues(interface{}) []interface{}
	GetInsertQueries(string, interface{}) []Query

	DefaultFallback(Context, interface{}, interface{}) (interface{}, error)
	Validate(Context, interface{}) (interface{}, error)
}
