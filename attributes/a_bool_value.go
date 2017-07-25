package attributes

import (
	"encoding/json"
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
func (bv BoolValue) Equal(obv BoolValue) bool {
	if bv.Value == nil {
		return (obv.Value == nil)
	} else if obv.Value == nil {
		return false
	}
	return (*bv.Value == *obv.Value)
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
