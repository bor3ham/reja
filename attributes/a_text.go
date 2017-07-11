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
	Default    func(interface{}) TextValue
}

func (t Text) GetKey() string {
	return t.Key
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

func (t *Text) DefaultFallback(val interface{}, instance interface{}) (interface{}, error) {
	if val == nil || !AssertText(val).Provided {
		if t.Default != nil {
			return t.Default(instance), nil
		}
		return nil, nil
	}
	return val, nil
}
func (t *Text) Validate(val interface{}) (interface{}, error) {
	textVal := AssertText(val)
	if textVal.Value == nil {
		if !t.Nullable {
			return textVal, errors.New(fmt.Sprintf("Attribute '%s' cannot be null.", t.Key))
		}
	} else {
		if t.MinLength != nil && len(*textVal.Value) < *t.MinLength {
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
		if t.MaxLength != nil && len(*textVal.Value) > *t.MaxLength {
			return nil, errors.New(fmt.Sprintf(
				"Attribute '%s' must be fewer than %d characters long.",
				t.Key,
				*t.MaxLength,
			))
		}
	}
	return textVal, nil
}

func (t *Text) GetInsertColumns(val interface{}) []string {
	var columns []string
	columns = append(columns, t.ColumnName)
	return columns
}
func (t *Text) GetInsertValues(val interface{}) []interface{} {
	textVal := AssertText(val)

	var values []interface{}
	values = append(values, textVal.Value)
	return values
}
