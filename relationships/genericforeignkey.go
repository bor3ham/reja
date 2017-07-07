package relationships

import (
	"github.com/bor3ham/reja/schema"
)

type GenericForeignKey struct {
	RelationshipStub
	Key            string
	TypeColumnName string
	IDColumnName   string
}

func (gfk GenericForeignKey) GetKey() string {
	return gfk.Key
}
func (gfk GenericForeignKey) GetType() string {
	return ""
}

func (gfk GenericForeignKey) GetSelectExtraColumns() []string {
	return []string{
		gfk.TypeColumnName,
		gfk.IDColumnName,
	}
}
func (gfk GenericForeignKey) GetSelectExtraVariables() []interface{} {
	var typeDest *string
	var idDest *string
	return []interface{}{
		&typeDest,
		&idDest,
	}
}

func (gfk GenericForeignKey) GetDefaultValue() interface{} {
	return &Pointer{}
}
func (gfk GenericForeignKey) GetValues(
	c schema.Context,
	m *schema.Model,
	ids []string,
	extra [][]interface{},
) (
	map[string]interface{},
	map[string]map[string][]string,
) {
	values := map[string]interface{}{}
	maps := map[string]map[string][]string{}
	for index, result := range extra {
		myId := ids[index]

		// parse extra columns
		modelType, ok := result[0].(**string)
		if !ok {
			panic("Unable to convert extra type")
		}
		stringId, ok := result[1].(**string)
		if !ok {
			panic("Unable to convert extra fk id")
		}

		// check value does not already exist
		// a foreign key can only have one value
		_, exists := values[myId]
		if exists {
			existingValue, ok := values[myId].(Pointer)
			if !ok {
				panic("Unable to convert previous value")
			}
			if *stringId == nil {
				if existingValue.Data != nil {
					panic("Contradictory values in query results")
				}
			} else {
				if existingValue.Data == nil ||
					*existingValue.Data.ID != **stringId ||
					existingValue.Data.Type != **modelType {
					panic("Contradictory values in query results")
				}
			}

			continue
		}

		var newValue Pointer
		if *stringId == nil {
			newValue = Pointer{}
		} else {
			newValue = Pointer{
				Data: &schema.InstancePointer{
					Type: **modelType,
					ID:   *stringId,
				},
			}
		}
		// update the value
		values[myId] = &newValue

		// add to relation map
		if *stringId != nil {
			_, exists = maps[myId]
			if !exists {
				maps[myId] = map[string][]string{}
			}
			_, exists = maps[myId][**modelType]
			if !exists {
				maps[myId][**modelType] = []string{}
			}
			maps[myId][**modelType] = append(maps[myId][**modelType], **stringId)
		}
	}

	return values, maps
}

func AssertGenericForeignKey(val interface{}) *Pointer {
	gfkVal, ok := val.(*Pointer)
	if !ok {
		panic("Bad generic foreign key value")
	}
	return gfkVal
}
