package relationships

import (
	"github.com/bor3ham/reja/context"
)

type ForeignKey struct {
	Key        string
	ColumnName string
	Type       string
}

func (fk ForeignKey) GetKey() string {
	return fk.Key
}
func (fk ForeignKey) GetType() string {
	return fk.Type
}

func (fk ForeignKey) GetInstanceColumnNames() []string {
	return []string{fk.ColumnName}
}
func (fk ForeignKey) GetInstanceColumnVariables() []interface{} {
	var destination *string
	return []interface{}{
		&destination,
	}
}
func (fk ForeignKey) GetExtraColumnNames() []string {
	return []string{fk.ColumnName}
}
func (fk ForeignKey) GetExtraColumnVariables() []interface{} {
	var destination *string
	return []interface{}{
		&destination,
	}
}

func (fk ForeignKey) GetDefaultValue() interface{} {
	return nil
}
func (fk ForeignKey) GetValues(c context.Context, ids []string, extra [][]interface{}) (map[string]interface{}, []string) {
	var relationIds []string
	stringId, ok := extra[0][0].(**string)
	if !ok {
		panic("Unable to convert extra fk id")
	}
	relationIds = append(relationIds, **stringId)

	return map[string]interface{}{}, relationIds
}
