package attributes

import (
	"encoding/json"
)

type TextValue struct {
	Value *string
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

type Text struct {
	ColumnName string
}

func (t Text) GetColumnNames() []string {
	return []string{t.ColumnName}
}
func (t Text) GetColumnVariables() []interface{} {
	var destination *string
	return []interface{}{
		&destination,
	}
}

func AssertText(val interface{}) TextValue {
	textVal, ok := val.(TextValue)
	if !ok {
		stringVal, ok := val.(**string)
		if !ok {
			panic("Bad text value")
		}
		return TextValue{
			Value: *stringVal,
			Provided: true,
		}
	}
	return textVal
}
