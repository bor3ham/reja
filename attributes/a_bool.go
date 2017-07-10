package attributes

import (
	"errors"
	"fmt"
)

type Bool struct {
	AttributeStub
	Key        string
	ColumnName string
	Nullable   bool
	Default    func(interface{}) BoolValue
}

func (b Bool) GetSelectDirectColumns() []string {
	return []string{b.ColumnName}
}
func (b Bool) GetSelectDirectVariables() []interface{} {
	var destination *bool
	return []interface{}{
		&destination,
	}
}

func (b *Bool) DefaultFallback(val interface{}, instance interface{}) (interface{}, error) {
	boolVal := AssertBool(val)
	if !boolVal.Provided {
		if b.Default != nil {
			return b.Default(instance), nil
		}
		return nil, nil
	}
	return boolVal, nil
}
func (b *Bool) Validate(val interface{}) (interface{}, error) {
	boolVal := AssertBool(val)
	if boolVal.Value == nil {
		if !b.Nullable {
			return nil, errors.New(fmt.Sprintf("Attribute '%s' cannot be null.", b.Key))
		}
	}
	return boolVal, nil
}

func (b *Bool) GetInsertColumns(val interface{}) []string {
	var columns []string
	columns = append(columns, b.ColumnName)
	return columns
}
func (b *Bool) GetInsertValues(val interface{}) []interface{} {
	boolVal := AssertBool(val)

	var values []interface{}
	values = append(values, boolVal.Value)
	return values
}
