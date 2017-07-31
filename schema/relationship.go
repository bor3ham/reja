package schema

type Relationship interface {
	GetKey() string
	GetType() string

	GetSelectExtraColumns() []string
	GetSelectExtraVariables() []interface{}

	AvailableFilters() []interface{}
	ValidateFilters(map[string][]string) ([]Filter, error)
	GetFilterWhere(int, map[string][]string) ([]string, []interface{})

	GetDefaultValue() interface{}
	GetValues(
		Context,
		*Model,
		[]string,
		[][]interface{},
		bool,
	) (
		map[string]interface{},
		map[string]map[string][]string,
	)

	GetInsert(interface{}) ([]string, []interface{})
	GetInsertQueries(string, interface{}) []Query
	GetUpdateQueries(string, interface{}, interface{}) []Query

	DefaultFallback(Context, interface{}, interface{}) (interface{}, error)
	Validate(Context, interface{}) (interface{}, error)
	ValidateUpdate(Context, interface{}, interface{}) (interface{}, error)
}
