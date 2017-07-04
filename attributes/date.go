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
	Default func(interface{}) DateValue
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

func (d *Date) DefaultFallback(val interface{}, instance interface{}) interface{} {
	dVal := AssertDate(val)
	if !dVal.Provided {
		if d.Default != nil {
			return d.Default(instance)
		}
		return nil
	}
	return dVal
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
