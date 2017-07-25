package attributes

import (
	"encoding/json"
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
func (iv IntegerValue) Equal(oiv IntegerValue) bool {
	if iv.Value == nil {
		return (oiv.Value == nil)
	} else if oiv.Value == nil {
		return false
	}
	return (*iv.Value == *oiv.Value)
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
