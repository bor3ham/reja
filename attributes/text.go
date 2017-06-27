package attributes

import (
	"encoding/json"
	"fmt"
	"errors"
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

type Text struct {
	Key string
	ColumnName string
	Nullable bool
	MinLength *int
	MaxLength *int
	Default *string
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
func (t *Text) ValidateNew(val interface{}) (interface{}, error) {
	textVal := AssertText(val)
	if !textVal.Provided && t.Default != nil {
		textVal.Value = t.Default
	}
	return t.validate(textVal)
}
func (t *Text) ValidateUpdate(val interface{}, oldVal interface{}) (interface{}, error) {
	return nil, nil
}
func (t *Text) validate(val TextValue) (interface{}, error) {
	if val.Value == nil {
		if !t.Nullable {
			return nil, errors.New(fmt.Sprintf("Attribute '%s' cannot be null.", t.Key))
		}
	} else {
		if t.MinLength != nil && len(*val.Value) < *t.MinLength {
			if *t.MinLength == 1 {
				return nil, errors.New(fmt.Sprintf(
					"Attribute '%s' cannot be blank.",
					t.Key,
				))
			}
			return nil, errors.New(fmt.Sprintf(
				"Attribute '%s' must be more than %d characters long.",
				t.Key,
				*t.MinLength,
			))
		}
		if t.MaxLength != nil && len(*val.Value) > *t.MaxLength {
			return nil, errors.New(fmt.Sprintf(
				"Attribute '%s' must be fewer than %d characters long.",
				t.Key,
				*t.MaxLength,
			))
		}
	}

	return val, nil
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
