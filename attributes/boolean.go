package attributes

import (
	"encoding/json"
	"errors"
	"fmt"
)

type BoolValue struct {
	Value    *bool
	Provided bool
}

func (bv *BoolValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(bv.Value)
}
func (bv *BoolValue) UnmarshalJSON(data []byte) error {
	bv.Provided = true

	if string(data) == "null" {
		return nil
	}

	var val bool
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	bv.Value = &val
	return nil
}

type Bool struct {
	Key        string
	ColumnName string
	Nullable   bool
	Default    *bool
}

func (b Bool) GetColumnNames() []string {
	return []string{b.ColumnName}
}
func (b Bool) GetColumnVariables() []interface{} {
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

func AssertBool(val interface{}) BoolValue {
	bVal, ok := val.(BoolValue)
	if !ok {
		plainVal, ok := val.(**bool)
		if !ok {
			panic("Bad boolean value")
		}
		return BoolValue{
			Value:    *plainVal,
			Provided: true,
		}
	}
	return bVal
}
