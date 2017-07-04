package attributes

type Attribute interface {
	GetSelectDirectColumns() []string
	GetSelectDirectVariables() []interface{}

	ValidateNew(interface{}) (interface{}, error)

	GetInsertColumns(interface{}) []string
	GetInsertValues(interface{}) []interface{}
}

type AttributeStub struct {}

func (stub AttributeStub) GetInsertColumns(val interface{}) []string {
	return []string{}
}
func (stub AttributeStub) GetInsertValues(val interface{}) []interface{} {
	return []interface{}{}
}
