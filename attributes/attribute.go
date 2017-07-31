package attributes

import (
	"github.com/bor3ham/reja/schema"
)

type AttributeStub struct{}

func (stub AttributeStub) GetOrderMap() map[string]string {
	return map[string]string{}
}

func (stub AttributeStub) AvailableFilters() []interface{} {
	return []interface{}{}
}
func (stub AttributeStub) ValidateFilters(map[string][]string) ([]schema.Filter, error) {
	return []schema.Filter{}, nil
}
func (stub AttributeStub) GetFilterWhere(int, map[string][]string) ([]string, []interface{}) {
	return []string{}, []interface{}{}
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
func (stub AttributeStub) ValidateUpdate(newVal interface{}, oldVal interface{}) (interface{}, error) {
	return nil, nil
}
func (stub AttributeStub) GetInsertColumns(val interface{}) []string {
	return []string{}
}
func (stub AttributeStub) GetInsertValues(val interface{}) []interface{} {
	return []interface{}{}
}
