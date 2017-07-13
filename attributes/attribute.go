package attributes

type AttributeStub struct{}

func (stub AttributeStub) ValidateFilters(map[string][]string) (map[string][]string, error) {
	return map[string][]string{}, nil
}
func (stub AttributeStub) GetFilterWhere(int, map[string][]string) ([]string, []interface{}) {
	return []string{}, []interface{}{}
}
func (stub AttributeStub) GetFilterAnnotate(map[string]string) string {
	return ""
}

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
