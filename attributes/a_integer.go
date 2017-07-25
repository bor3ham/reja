package attributes

import (
	"errors"
	"fmt"
)

type Integer struct {
	AttributeStub
	Key        string
	ColumnName string
	Nullable   bool
	Default    func(interface{}) IntegerValue
}

func (i Integer) GetKey() string {
	return i.Key
}

func (i Integer) GetSelectDirectColumns() []string {
	return []string{i.ColumnName}
}
func (i Integer) GetSelectDirectVariables() []interface{} {
	var destination *int
	return []interface{}{
		&destination,
	}
}

func (i *Integer) DefaultFallback(val interface{}, instance interface{}) (interface{}, error) {
	iVal := AssertInteger(val)
	if !iVal.Provided {
		if i.Default != nil {
			return i.Default(instance), nil
		}
		return nil, nil
	}
	return iVal, nil
}
func (i *Integer) Validate(val interface{}) (interface{}, error) {
	iVal := AssertInteger(val)
	if iVal.Value == nil {
		if !i.Nullable {
			return nil, errors.New(fmt.Sprintf("Attribute '%s' cannot be null.", i.Key))
		}
	}
	return iVal, nil
}
func (i *Integer) ValidateUpdate(newVal interface{}, oldVal interface{}) (interface{}, error) {
	newInteger := AssertInteger(newVal)
	oldInteger := AssertInteger(oldVal)
	if !newInteger.Provided {
		return nil, nil
	}
	valid, err := i.Validate(newInteger)
	if err != nil {
		return nil, err
	}
	validNewInteger := AssertInteger(valid)
	if validNewInteger.Equal(oldInteger) {
		return nil, nil
	}
	return validNewInteger, nil
}

func (i *Integer) GetInsertColumns(val interface{}) []string {
	var columns []string
	columns = append(columns, i.ColumnName)
	return columns
}
func (i *Integer) GetInsertValues(val interface{}) []interface{} {
	iVal := AssertInteger(val)

	var values []interface{}
	values = append(values, iVal.Value)
	return values
}
