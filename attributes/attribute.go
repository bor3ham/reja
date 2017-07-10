package attributes

type AttributeStub struct{}

func (stub AttributeStub) DefaultFallback(
	val interface{},
	instance interface{},
) (
	interface{},
	error,
) {
	return val, nil
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
