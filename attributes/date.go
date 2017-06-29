package attributes

import (
	"fmt"
	"time"
	"strings"
	"errors"
)

const DATE_LAYOUT = "2006-01-02"

type DateValue struct {
	Value *time.Time
	Provided bool
}

func (dv DateValue) MarshalJSON() ([]byte, error) {
	var stamp string
	if dv.Value == nil {
		stamp = "\"null\""
	} else {
		stamp = fmt.Sprintf("\"%s\"", time.Time(*dv.Value).Format(DATE_LAYOUT))
	}
	return []byte(stamp), nil
}
func (dv *DateValue) UnmarshalJSON(data []byte) error {
	dv.Provided = true

	strData := string(data)
	if strData == "null" {
		return nil
	}

	val, err := time.Parse(DATE_LAYOUT, strings.Trim(strData, "\""))
	if err != nil {
		return err
	}
	dv.Value = &val
	return nil
}

type Date struct {
	Key string
	ColumnName string
	Default *time.Time
	Nullable bool
}

func (d Date) GetColumnNames() []string {
	return []string{d.ColumnName}
}
func (d Date) GetColumnVariables() []interface{} {
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

func AssertDate(val interface{}) DateValue {
	dVal, ok := val.(DateValue)
	if !ok {
		plainVal, ok := val.(**time.Time)
		if !ok {
			panic("Bad date value")
		}
		return DateValue{
			Value: *plainVal,
			Provided: true,
		}
	}
	return dVal
}
