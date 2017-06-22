package relationships

import (
	"github.com/bor3ham/reja/context"
)

type GenericForeignKey struct {
	Key        string
	TypeColumnName string
	IDColumnName string
}

func (gfk GenericForeignKey) GetKey() string {
	return gfk.Key
}
func (gfk GenericForeignKey) GetType() string {
	return ""
}

func (gfk GenericForeignKey) GetInstanceColumnNames() []string {
	return []string{
		gfk.TypeColumnName,
		gfk.IDColumnName,
	}
}
func (gfk GenericForeignKey) GetInstanceColumnVariables() []interface{} {
	var typeDest *string
	var idDest *string
	return []interface{}{
		&typeDest,
		&idDest,
	}
}
func (gfk GenericForeignKey) GetExtraColumnNames() []string {
	return []string{
		gfk.TypeColumnName,
		gfk.IDColumnName,
	}
}
func (gfk GenericForeignKey) GetExtraColumnVariables() []interface{} {
	var typeDest *string
	var idDest *string
	return []interface{}{
		&typeDest,
		&idDest,
	}
}

func (gfk GenericForeignKey) GetDefaultValue() interface{} {
	return nil
}
func (gfk GenericForeignKey) GetValues(
	c context.Context,
	ids []string,
	extra [][]interface{},
) (
	map[string]interface{},
	map[string][]string,
) {
	relationMap := map[string][]string{}
	for _, result := range extra {
		stringType, ok := result[0].(**string)
		if !ok {
			panic("Unable to convert extra gfk type")
		}
		stringId, ok := result[1].(**string)
		if !ok {
			panic("Unable to convert extra gfk id")
		}
		_, exists := relationMap[**stringType]
		if !exists {
			relationMap[**stringType] = []string{}
		}
		relationMap[**stringType] = append(relationMap[**stringType], **stringId)
	}

	return map[string]interface{}{}, relationMap
}
