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
	Default    *bool
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
func (b *Bool) ValidateNew(val interface{}) (interface{}, error) {
	boolVal := AssertBool(val)
	if !boolVal.Provided && b.Default != nil {
		boolVal.Value = b.Default
	}
	return b.validate(boolVal)
}
func (b *Bool) validate(val BoolValue) (interface{}, error) {
	if val.Value == nil {
		if !b.Nullable {
			return nil, errors.New(fmt.Sprintf("Attribute '%s' cannot be null.", b.Key))
		}
	}
	return val, nil
}
