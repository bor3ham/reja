package schema

type Filter interface {
	GetQArgKey() string
	GetQArgValues() []string

	GetWhereQueries(int) []string
	GetWhereArgs() []interface{}
}
