package relationships

import (
	"github.com/bor3ham/reja/context"
	"github.com/bor3ham/reja/instances"
)

type ForeignKey struct {
	RelationshipStub
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

func (fk ForeignKey) GetSelectExtraColumns() []string {
	return []string{fk.ColumnName}
}
func (fk ForeignKey) GetSelectExtraVariables() []interface{} {
	var destination *string
	return []interface{}{
		&destination,
	}
}

func (fk ForeignKey) GetDefaultValue() interface{} {
	return &Pointer{}
}
func (fk ForeignKey) GetValues(
	c context.Context,
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
		stringId, ok := result[0].(**string)
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
					existingValue.Data.Type != fk.Type {
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
				Data: &instances.InstancePointer{
					Type: fk.Type,
					ID:   *stringId,
				},
			}
		}
		values[myId] = &newValue

		// add to relation map
		if *stringId != nil {
			_, exists = maps[myId]
			if !exists {
				maps[myId] = map[string][]string{}
			}
			_, exists = maps[myId][fk.Type]
			if !exists {
				maps[myId][fk.Type] = []string{}
			}
			maps[myId][fk.Type] = append(maps[myId][fk.Type], **stringId)
		}
	}

	return values, maps
}

func AssertForeignKey(val interface{}) *Pointer {
	fkVal, ok := val.(*Pointer)
	if !ok {
		panic("Bad foreign key value")
	}
	return fkVal
}
