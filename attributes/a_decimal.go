package attributes

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
)

type Decimal struct {
	AttributeStub
	Key           string
	ColumnName    string
	DecimalPlaces int32
	Nullable      bool
	Default       func(interface{}) DecimalValue
}

func (d Decimal) GetKey() string {
	return d.Key
}

func (d Decimal) GetSelectDirect() ([]string, []interface{}) {
	var destination *decimal.Decimal
	return []string{d.ColumnName}, []interface{}{
		&destination,
	}
}

func (d Decimal) GetOrderMap() map[string]string {
	orders := map[string]string{}
	orders[d.Key] = d.ColumnName
	return orders
}

func (d *Decimal) DefaultFallback(val interface{}, instance interface{}) (interface{}, error) {
	dVal := AssertDecimal(val, d.DecimalPlaces)
	if !dVal.Provided {
		if d.Default != nil {
			return d.Default(instance), nil
		}
		return nil, nil
	}
	return dVal, nil
}
func (d *Decimal) Validate(val interface{}) (interface{}, error) {
	dVal := AssertDecimal(val, d.DecimalPlaces)
	if dVal.Value == nil {
		if !d.Nullable {
			return nil, errors.New(fmt.Sprintf("Attribute '%s' cannot be null.", d.Key))
		}
	} else {
		truncValue := dVal.Value.Truncate(d.DecimalPlaces)
		dVal.Value = &truncValue
	}
	return dVal, nil
}
func (d *Decimal) ValidateUpdate(newVal interface{}, oldVal interface{}) (interface{}, error) {
	newDecimal := AssertDecimal(newVal, d.DecimalPlaces)
	oldDecimal := AssertDecimal(oldVal, d.DecimalPlaces)
	if !newDecimal.Provided {
		return nil, nil
	}
	valid, err := d.Validate(newDecimal)
	if err != nil {
		return nil, err
	}
	validNewDecimal := AssertDecimal(valid, d.DecimalPlaces)
	if validNewDecimal.Equal(oldDecimal) {
		return nil, nil
	}
	return validNewDecimal, nil
}

func (d *Decimal) GetInsert(val interface{}) ([]string, []interface{}) {
	dVal := AssertDecimal(val, d.DecimalPlaces)

	columns := []string{d.ColumnName}
	values := []interface{}{dVal.Value}
	return columns, values
}
