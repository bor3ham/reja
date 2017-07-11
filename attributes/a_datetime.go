package attributes

import (
	"time"
)

type Datetime struct {
	AttributeStub
	Key        string
	ColumnName string
}

func (dt Datetime) GetKey() string {
	return dt.Key
}

func (dt Datetime) GetSelectDirectColumns() []string {
	return []string{dt.ColumnName}
}
func (dt Datetime) GetSelectDirectVariables() []interface{} {
	var destination *time.Time
	return []interface{}{
		&destination,
	}
}
