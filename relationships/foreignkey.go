package relationships

import (
	"github.com/bor3ham/reja/context"
)

type ForeignKey struct {
	ColumnName string
	Type       string
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
