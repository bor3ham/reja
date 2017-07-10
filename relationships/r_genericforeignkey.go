package relationships

import (
	"github.com/bor3ham/reja/schema"
	"errors"
	"fmt"
)

type GenericForeignKey struct {
	RelationshipStub
	Key            string
	TypeColumnName string
	IDColumnName   string
	Nullable   bool
	Default    func(schema.Context, interface{}) Pointer
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
	return schema.Result{}
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

		var newValue schema.Result
		if *stringId == nil {
			newValue = schema.Result{}
		} else {
			newValue = schema.Result{
				Data: schema.InstancePointer{
					Type: **modelType,
					ID:   *stringId,
				},
			}
		}
		// update the value
		values[myId] = newValue

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

func (gfk *GenericForeignKey) DefaultFallback(
	c schema.Context,
	val interface{},
	instance interface{},
) (
	interface{},
	error,
) {
	gfkVal, err := ParseResultPointer(val)
	if err != nil {
		return nil, err
	}
	if !gfkVal.Provided {
		if gfk.Default != nil {
			return gfk.Default(c, instance), nil
		}
		return nil, nil
	}
	return gfkVal, nil
}
func (gfk *GenericForeignKey) Validate(c schema.Context, val interface{}) (interface{}, error) {
	gfkVal := AssertPointer(val)

	if gfkVal.Data == nil {
		if !gfk.Nullable {
			return nil, errors.New(fmt.Sprintf(
				"Relationship '%s' invalid: Cannot be null.",
				gfk.Key,
			))
		}
		return gfkVal, nil
	}

	valType := gfkVal.Data.Type
	if gfkVal.Data.ID == nil {
		return nil, errors.New(fmt.Sprintf(
			"Relationship '%s' invalid: Missing ID.",
			gfk.Key,
		))
	}
	valID := *gfkVal.Data.ID

	// check that the object exists
	model := c.GetServer().GetModel(valType)
	// validate the type exists
	if model == nil {
		return nil, errors.New(fmt.Sprintf(
			"Relationship '%s' invalid: Non existent type.",
			gfk.Key,
		))
	}
	include := schema.Include{
		Children: map[string]*schema.Include{},
	}
	instances, _, err := c.GetObjects(
		model,
		[]string{valID},
		0,
		0,
		&include,
	)
	if err != nil {
		panic(err)
	}
	if len(instances) == 0 {
		return nil, errors.New(fmt.Sprintf(
			"Relationship '%s' invalid: %s ID '%s' does not exist.",
			gfk.Key,
			valType,
			valID,
		))
	}
	return gfkVal, nil
}

func (gfk *GenericForeignKey) GetInsertColumns(val interface{}) []string {
	return []string{
		gfk.TypeColumnName,
		gfk.IDColumnName,
	}
}
func (gfk *GenericForeignKey) GetInsertValues(val interface{}) []interface{} {
	resultVal := AssertPointer(val)
	if resultVal.Data == nil {
		return []interface{}{
			nil,
			nil,
		}
	}
	return []interface{}{
		resultVal.Data.Type,
		resultVal.Data.ID,
	}
}

func AssertGenericForeignKey(val interface{}) schema.Result {
	gfkVal, ok := val.(schema.Result)
	if !ok {
		panic("Bad generic foreign key value")
	}
	return gfkVal
}
