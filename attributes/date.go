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
	Default    *time.Time
	Nullable   bool
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
func (d *Date) ValidateNew(val interface{}) (interface{}, error) {
	dVal := AssertDate(val)
	if !dVal.Provided && d.Default != nil {
		dVal.Value = d.Default
	}
	return d.validate(dVal)
}
func (d *Date) ValidateUpdate(newVal interface{}, oldVal interface{}) (interface{}, error) {
	newDVal := AssertDate(newVal)
	oldDVal := AssertDate(oldVal)
	if !newDVal.Provided {
		return oldDVal, nil
	}
	return d.validate(newDVal)
}
func (d *Date) validate(val DateValue) (interface{}, error) {
	if val.Value == nil {
		if !d.Nullable {
			return nil, errors.New(fmt.Sprintf("Attribute '%s' cannot be null.", d.Key))
		}
	}
	return val, nil
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
