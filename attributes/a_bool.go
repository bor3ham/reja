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

func (b Bool) GetKey() string {
	return b.Key
}

func (b Bool) GetSelectDirect() ([]string, []interface{}) {
	var destination *bool
	return []string{b.ColumnName}, []interface{}{
		&destination,
	}
}

func (b Bool) GetOrderMap() map[string]string {
	orders := map[string]string{}
	orders[b.Key] = b.ColumnName
	return orders
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
func (b *Bool) ValidateUpdate(newVal interface{}, oldVal interface{}) (interface{}, error) {
	newBool := AssertBool(newVal)
	oldBool := AssertBool(oldVal)
	if !newBool.Provided {
		return nil, nil
	}
	valid, err := b.Validate(newBool)
	if err != nil {
		return nil, err
	}
	validNewBool := AssertBool(valid)
	if validNewBool.Equal(oldBool) {
		return nil, nil
	}
	return validNewBool, nil
}

func (b *Bool) GetInsert(val interface{}) ([]string, []interface{}) {
	boolVal := AssertBool(val)

	columns := []string{b.ColumnName}
	values := []interface{}{boolVal.Value}
	return columns, values
}
