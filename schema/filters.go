package schema

type Filter interface {
	GetQArgKey() string
	GetQArgValues() []string

	GetWhereQueries(Context, int) []string
	GetWhereArgs() []interface{}
}

type BaseFilter struct {
	QArgKey    string
	QArgValues []string
}

func (bf BaseFilter) GetQArgKey() string {
	return bf.QArgKey
}
func (bf BaseFilter) GetQArgValues() []string {
	return bf.QArgValues
}
