package attributes

import (
	"fmt"
	"time"
)

type DateValue time.Time

func (dv DateValue) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(dv).Format("2006-01-02"))
	return []byte(stamp), nil
}

type Date struct {
	ColumnName string
}

func (d Date) GetColumnNames() []string {
	return []string{d.ColumnName}
}
func (d Date) GetColumnVariables() []interface{} {
	var destination *DateValue
	return []interface{}{
		&destination,
	}
}
func (d *Date) ValidateNew(val interface{}) error {
	return nil
}

func AssertDate(val interface{}) *DateValue {
	dVal, ok := val.(**DateValue)
	if !ok {
		panic("Bad date value")
	}
	return *dVal
}
