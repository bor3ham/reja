package attributes

import (
	"errors"
	"fmt"
	"strings"
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

func (t Text) GetSelectDirect() ([]string, []interface{}) {
	var destination *string
	return []string{t.ColumnName}, []interface{}{
		&destination,
	}
}

func (t Text) GetOrderMap() map[string]string {
	orders := map[string]string{}
	orders[t.Key] = t.ColumnName
	return orders
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
		trimmedValue := strings.TrimSpace(*textVal.Value)
		textVal.Value = &trimmedValue
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
func (t *Text) ValidateUpdate(newVal interface{}, oldVal interface{}) (interface{}, error) {
	newText := AssertText(newVal)
	oldText := AssertText(oldVal)
	if !newText.Provided {
		return nil, nil
	}
	valid, err := t.Validate(newText)
	if err != nil {
		return nil, err
	}
	validNewText := AssertText(valid)
	if validNewText.Equal(oldText) {
		return nil, nil
	}
	return validNewText, nil
}

func (t *Text) GetInsert(val interface{}) ([]string, []interface{}) {
	textVal := AssertText(val)

	columns := []string{t.ColumnName}
	values := []interface{}{textVal.Value}
	return columns, values
}
