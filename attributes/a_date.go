package attributes

import (
	"errors"
	"fmt"
	"time"
)

const DATE_LAYOUT = "2006-01-02"

type Date struct {
	AttributeStub
	Key        string
	ColumnName string
	Nullable   bool
	Default    func(interface{}) DateValue
}

func (d Date) GetKey() string {
	return d.Key
}

func (d Date) GetSelectDirectColumns() []string {
	return []string{d.ColumnName}
}
func (d Date) GetSelectDirectVariables() []interface{} {
	var destination *time.Time
	return []interface{}{
		&destination,
	}
}

func (d *Date) DefaultFallback(val interface{}, instance interface{}) (interface{}, error) {
	dVal := AssertDate(val)
	if !dVal.Provided {
		if d.Default != nil {
			return d.Default(instance), nil
		}
		return nil, nil
	}
	return dVal, nil
}
func (d *Date) Validate(val interface{}) (interface{}, error) {
	dVal := AssertDate(val)
	if dVal.Value == nil {
		if !d.Nullable {
			return nil, errors.New(fmt.Sprintf("Attribute '%s' cannot be null.", d.Key))
		}
	}
	return dVal, nil
}
func (d *Date) ValidateUpdate(newVal interface{}, oldVal interface{}) (interface{}, error) {
	newDate := AssertDate(newVal)
	oldDate := AssertDate(oldVal)
	if !newDate.Provided {
		return nil, nil
	}
	valid, err := d.Validate(newDate)
	if err != nil {
		return nil, err
	}
	validNewDate := AssertDate(valid)
	if validNewDate.Equal(oldDate) {
		return nil, nil
	}
	return validNewDate, nil
}

func (d *Date) GetInsertColumns(val interface{}) []string {
	var columns []string
	columns = append(columns, d.ColumnName)
	return columns
}
func (d *Date) GetInsertValues(val interface{}) []interface{} {
	dateVal := AssertDate(val)

	var values []interface{}
	values = append(values, dateVal.Value)
	return values
}
