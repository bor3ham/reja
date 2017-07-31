package attributes

import (
	"errors"
	"fmt"
	"time"
)

type Datetime struct {
	AttributeStub
	Key        string
	ColumnName string
	Nullable   bool
	Default    func(interface{}) DatetimeValue
}

func (dt Datetime) GetKey() string {
	return dt.Key
}

func (dt Datetime) GetSelectDirect() ([]string, []interface{}) {
	var destination *time.Time
	return []string{dt.ColumnName}, []interface{}{
		&destination,
	}
}

func (dt Datetime) GetOrderMap() map[string]string {
	orders := map[string]string{}
	orders[dt.Key] = dt.ColumnName
	return orders
}

func (dt *Datetime) DefaultFallback(val interface{}, instance interface{}) (interface{}, error) {
	dtVal := AssertDatetime(val)
	if !dtVal.Provided {
		if dt.Default != nil {
			return dt.Default(instance), nil
		}
		return nil, nil
	}
	return dtVal, nil
}
func (dt *Datetime) Validate(val interface{}) (interface{}, error) {
	dtVal := AssertDatetime(val)
	if dtVal.Value == nil {
		if !dt.Nullable {
			return nil, errors.New(fmt.Sprintf("Attribute '%s' cannot be null.", dt.Key))
		}
	}
	return dtVal, nil
}
func (dt *Datetime) ValidateUpdate(newVal interface{}, oldVal interface{}) (interface{}, error) {
	newDatetime := AssertDatetime(newVal)
	oldDatetime := AssertDatetime(oldVal)
	if !newDatetime.Provided {
		return nil, nil
	}
	valid, err := dt.Validate(newDatetime)
	if err != nil {
		return nil, err
	}
	validNewDatetime := AssertDatetime(valid)
	if validNewDatetime.Equal(oldDatetime) {
		return nil, nil
	}
	return validNewDatetime, nil
}

func (dt *Datetime) GetInsert(val interface{}) ([]string, []interface{}) {
	dtVal := AssertDatetime(val)

	columns := []string{dt.ColumnName}
	values := []interface{}{dtVal.Value}
	return columns, values
}
