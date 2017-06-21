package relationships

import (
	"github.com/bor3ham/reja/context"
)

type ForeignKey struct {
	Key string
	ColumnName string
	Type       string
}

func (fk ForeignKey) GetKey() string {
	return fk.Key
}
func (fk ForeignKey) GetType() string {
	return fk.Type
}

func (fk ForeignKey) GetColumnNames() []string {
	return []string{fk.ColumnName}
}
func (fk ForeignKey) GetColumnVariables() []interface{} {
	var destination *string
	return []interface{}{
		&destination,
	}
}

func (fk ForeignKey) GetDefaultValue() interface{} {
	return nil
}
func (fk ForeignKey) GetValues(c context.Context, ids []string) map[string]interface{} {
	return map[string]interface{}{}
}
