package attributes

import (
	"encoding/json"
	"github.com/shopspring/decimal"
)

type DecimalValue struct {
	Value         *decimal.Decimal
	DecimalPlaces int32
	Provided      bool
}

func (dv *DecimalValue) MarshalJSON() ([]byte, error) {
	if dv.Value == nil {
		return []byte("null"), nil
	}
	return []byte("\"" + dv.Value.StringFixed(dv.DecimalPlaces) + "\""), nil
}
func (dv *DecimalValue) UnmarshalJSON(data []byte) error {
	dv.Provided = true

	if string(data) == "null" {
		return nil
	}

	var val decimal.Decimal
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	dv.Value = &val
	return nil
}

func AssertDecimal(val interface{}, decimalPlaces int32) DecimalValue {
	dVal, ok := val.(DecimalValue)
	if !ok {
		plainVal, ok := val.(**decimal.Decimal)
		if !ok {
			panic("Bad decimal value")
		}
		return DecimalValue{
			Value:         *plainVal,
			DecimalPlaces: decimalPlaces,
			Provided:      true,
		}
	}
	dVal.DecimalPlaces = decimalPlaces
	return dVal
}
