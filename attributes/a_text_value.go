package attributes

import (
	"encoding/json"
)

type TextValue struct {
	Value    *string
	Provided bool
}

func (tv *TextValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(tv.Value)
}
func (tv *TextValue) UnmarshalJSON(data []byte) error {
	tv.Provided = true

	if string(data) == "null" {
		return nil
	}

	var val string
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	tv.Value = &val
	return nil
}
func (tv TextValue) Equal(otv TextValue) bool {
	if tv.Value == nil {
		return (otv.Value == nil)
	} else if otv.Value == nil {
		return false
	}
	return (*tv.Value == *otv.Value)
}

func AssertText(val interface{}) TextValue {
	textVal, ok := val.(TextValue)
	if !ok {
		plainVal, ok := val.(**string)
		if !ok {
			panic("Bad text value")
		}
		return TextValue{
			Value:    *plainVal,
			Provided: true,
		}
	}
	return textVal
}
