package attributes

type Attribute interface {
	GetColumnNames() []string
	GetColumnVariables() []interface{}
}
