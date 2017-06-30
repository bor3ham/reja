package attributes

import (
	"time"
	"fmt"
	"strings"
)

type DateValue struct {
	Value    *time.Time
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

func AssertDate(val interface{}) DateValue {
	dVal, ok := val.(DateValue)
	if !ok {
		plainVal, ok := val.(**time.Time)
		if !ok {
			panic("Bad date value")
		}
		return DateValue{
			Value:    *plainVal,
			Provided: true,
		}
	}
	return dVal
}
