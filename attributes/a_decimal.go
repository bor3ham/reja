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

func (d Decimal) GetSelectDirectColumns() []string {
	return []string{d.ColumnName}
}
func (d Decimal) GetSelectDirectVariables() []interface{} {
	var destination *decimal.Decimal
	return []interface{}{
		&destination,
	}
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
	}
	return dVal, nil
}

func (d *Decimal) GetInsertColumns(val interface{}) []string {
	var columns []string
	columns = append(columns, d.ColumnName)
	return columns
}
func (d *Decimal) GetInsertValues(val interface{}) []interface{} {
	dVal := AssertDecimal(val, d.DecimalPlaces)

	var values []interface{}
	values = append(values, dVal.Value)
	return values
}
