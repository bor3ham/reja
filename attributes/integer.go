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
	Default    *int
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
func (i *Integer) ValidateNew(val interface{}) (interface{}, error) {
	intVal := AssertInteger(val)
	if !intVal.Provided && i.Default != nil {
		intVal.Value = i.Default
	}
	return i.validate(intVal)
}
func (i *Integer) validate(val IntegerValue) (interface{}, error) {
	if val.Value == nil {
		if !i.Nullable {
			return nil, errors.New(fmt.Sprintf("Attribute '%s' cannot be null.", i.Key))
		}
	}
	return val, nil
}
