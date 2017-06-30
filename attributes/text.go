package attributes

import (
	"errors"
	"fmt"
)

type Text struct {
	AttributeStub
	Key        string
	ColumnName string
	Nullable   bool
	MinLength  *int
	MaxLength  *int
	Default    *string
}

func (t Text) GetSelectDirectColumns() []string {
	return []string{t.ColumnName}
}
func (t Text) GetSelectDirectVariables() []interface{} {
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

func (t *Text) GetInsertColumns() []string {
	return []string{}
}
func (t *Text) GetInsertValues() []interface{} {
	return []interface{}{}
}
