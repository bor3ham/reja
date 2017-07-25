package attributes

import (
	"fmt"
	"strings"
	"time"
)

type DatetimeValue struct {
	Value    *time.Time
	Provided bool
}

func (dtv DatetimeValue) MarshalJSON() ([]byte, error) {
	var stamp string
	if dtv.Value == nil {
		stamp = "null"
	} else {
		stamp = fmt.Sprintf("\"%s\"", time.Time(*dtv.Value).Format(time.RFC3339))
	}
	return []byte(stamp), nil
}
func (dtv *DatetimeValue) UnmarshalJSON(data []byte) error {
	dtv.Provided = true

	strData := string(data)
	if strData == "null" {
		return nil
	}

	val, err := time.Parse(time.RFC3339, strings.Trim(strData, "\""))
	if err != nil {
		return err
	}
	dtv.Value = &val
	return nil
}
func (dtv DatetimeValue) Equal(odtv DatetimeValue) bool {
	if dtv.Value == nil {
		return (odtv.Value == nil)
	} else if odtv.Value == nil {
		return false
	}
	return (dtv.Value.Equal(*odtv.Value))
}

func AssertDatetime(val interface{}) DatetimeValue {
	dtVal, ok := val.(DatetimeValue)
	if !ok {
		plainVal, ok := val.(**time.Time)
		if !ok {
			panic("Bad datetime value")
		}
		return DatetimeValue{
			Value:    *plainVal,
			Provided: true,
		}
	}
	return dtVal
}
