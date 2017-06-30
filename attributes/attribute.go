package attributes

type Attribute interface {
	GetSelectDirectColumns() []string
	GetSelectDirectVariables() []interface{}

	ValidateNew(interface{}) (interface{}, error)

	GetInsertColumns() []string
	GetInsertValues() []interface{}
}

type AttributeStub struct {}

func (stub AttributeStub) GetInsertColumns() []string {
	return []string{}
}
func (stub AttributeStub) GetInsertValues() []interface{} {
	return []interface{}{}
}
