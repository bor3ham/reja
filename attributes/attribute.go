package attributes

type Attribute interface {
	GetSelectDirectColumns() []string
	GetSelectDirectVariables() []interface{}

	DefaultFallback(interface{}, interface{}) interface{}
	Validate(interface{}) (interface{}, error)

	GetInsertColumns(interface{}) []string
	GetInsertValues(interface{}) []interface{}
}

type AttributeStub struct{}

func (stub AttributeStub) DefaultFallback(val interface{}, instance interface{}) interface{} {
	return val
}
func (stub AttributeStub) Validate(val interface{}) (interface{}, error) {
	return val, nil
}
func (stub AttributeStub) GetInsertColumns(val interface{}) []string {
	return []string{}
}
func (stub AttributeStub) GetInsertValues(val interface{}) []interface{} {
	return []interface{}{}
}
