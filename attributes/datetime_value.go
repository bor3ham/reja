package attributes

import (
	"time"
	"fmt"
)

type DatetimeValue time.Time

func (dtv DatetimeValue) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(dtv).Format(time.RFC3339))
	return []byte(stamp), nil
}

func AssertDatetime(val interface{}) *DatetimeValue {
	dtVal, ok := val.(**DatetimeValue)
	if !ok {
		panic("Bad datetime value")
	}
	return *dtVal
}
