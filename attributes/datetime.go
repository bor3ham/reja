package attributes

import (
	"fmt"
	"time"
)

type DatetimeValue time.Time

func (dtv DatetimeValue) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(dtv).Format(time.RFC3339))
	return []byte(stamp), nil
}

type Datetime struct {
	ColumnName string
}

func (dt Datetime) GetColumnNames() []string {
	return []string{dt.ColumnName}
}
func (dt Datetime) GetColumnVariables() []interface{} {
	var destination *DatetimeValue
	return []interface{}{
		&destination,
	}
}
func (dt *Datetime) ValidateNew(val interface{}) (interface{}, error) {
	return nil, nil
}

func AssertDatetime(val interface{}) *DatetimeValue {
	dtVal, ok := val.(**DatetimeValue)
	if !ok {
		panic("Bad datetime value")
	}
	return *dtVal
}
