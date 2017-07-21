package schema

type Filter interface {
	GetQArgKey() string
	GetQArgValues() []string

	GetWhere(Context, string, int) ([]string, []interface{})
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
