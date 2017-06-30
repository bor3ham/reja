package attributes

import (
	"encoding/json"
	"errors"
	"fmt"
)

type IntegerValue struct {
	Value    *int
	Provided bool
}

func (iv *IntegerValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(iv.Value)
}
func (iv *IntegerValue) UnmarshalJSON(data []byte) error {
	iv.Provided = true

	if string(data) == "null" {
		return nil
	}

	var val int
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	iv.Value = &val
	return nil
}

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

func AssertInteger(val interface{}) IntegerValue {
	intVal, ok := val.(IntegerValue)
	if !ok {
		plainVal, ok := val.(**int)
		if !ok {
			panic("Bad integer value")
		}
		return IntegerValue{
			Value:    *plainVal,
			Provided: true,
		}
	}
	return intVal
}
