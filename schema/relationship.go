package schema

type Relationship interface {
	GetKey() string
	GetType() string

	GetSelectDirectColumns() []string
	GetSelectDirectVariables() []interface{}
	GetSelectExtraColumns() []string
	GetSelectExtraVariables() []interface{}

	GetDefaultValue() interface{}
	GetValues(
		Context,
		[]string,
		[][]interface{},
	) (
		map[string]interface{},
		map[string]map[string][]string,
	)

	GetInsertQueries(string, interface{}) []Query

	DefaultFallback(Context, interface{}, interface{}) interface{}
	Validate(Context, interface{}) (interface{}, error)
}
