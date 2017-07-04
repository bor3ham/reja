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
	Default func(interface{}) IntegerValue
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

func (i *Integer) DefaultFallback(val interface{}, instance interface{}) interface{} {
	iVal := AssertInteger(val)
	if !iVal.Provided {
		if i.Default != nil {
			return i.Default(instance)
		}
		return nil
	}
	return iVal
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
