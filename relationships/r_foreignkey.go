package relationships

import (
	"errors"
	"fmt"
	"github.com/bor3ham/reja/schema"
)

type ForeignKey struct {
	RelationshipStub
	Key        string
	ColumnName string
	Type       string
	Nullable   bool
	Default    func(schema.Context, interface{}) Pointer
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
	return schema.Result{}
}
func (fk ForeignKey) GetValues(
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

		selfLink := relationLink(c, m.Type, myId, fk.Key)
		newValue := schema.Result{
			Links: map[string]*string{
				"self": &selfLink,
			},
		}
		if *stringId != nil {
			newValue.Data = schema.InstancePointer{
				Type: fk.Type,
				ID:   *stringId,
			}
		}
		values[myId] = newValue

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

func (fk *ForeignKey) DefaultFallback(
	c schema.Context,
	val interface{},
	instance interface{},
) (
	interface{},
	error,
) {
	var fkVal Pointer
	if val != nil {
		fkVal = Pointer{Provided: false}
	} else {
		var err error
		fkVal, err = ParseResultPointer(val)
		if err != nil {
			return nil, err
		}
	}

	if !fkVal.Provided {
		if fk.Default != nil {
			return fk.Default(c, instance), nil
		}
		return nil, nil
	}
	return fkVal, nil
}
func (fk *ForeignKey) Validate(c schema.Context, val interface{}) (interface{}, error) {
	fkVal := AssertPointer(val)

	if fkVal.Data == nil {
		if !fk.Nullable {
			return nil, errors.New(fmt.Sprintf(
				"Relationship '%s' invalid: Cannot be null.",
				fk.Key,
			))
		}
		return fkVal, nil
	}

	valType := fkVal.Data.Type
	if fkVal.Data.ID == nil {
		return nil, errors.New(fmt.Sprintf(
			"Relationship '%s' invalid: Missing ID.",
			fk.Key,
		))
	}
	valID := *fkVal.Data.ID

	// validate the type is correct
	if valType != fk.Type {
		return nil, errors.New(fmt.Sprintf(
			"Relationship '%s' invalid: Incorrect type.",
			fk.Key,
		))
	}

	// check that the object exists
	model := c.GetServer().GetModel(fk.Type)
	include := schema.Include{
		Children: map[string]*schema.Include{},
	}
	instances, _, err := c.GetObjectsByIDs(model, []string{valID}, &include)
	if err != nil {
		panic(err)
	}
	if len(instances) == 0 {
		return nil, errors.New(fmt.Sprintf(
			"Relationship '%s' invalid: %s ID '%s' does not exist.",
			fk.Key,
			fk.Type,
			valID,
		))
	}
	return fkVal, nil
}

func (fk *ForeignKey) GetInsertColumns(val interface{}) []string {
	return []string{
		fk.ColumnName,
	}
}
func (fk *ForeignKey) GetInsertValues(val interface{}) []interface{} {
	resultVal := AssertPointer(val)
	if resultVal.Data == nil {
		return []interface{}{
			nil,
		}
	}
	return []interface{}{
		resultVal.Data.ID,
	}
}

func AssertForeignKey(val interface{}) schema.Result {
	fkVal, ok := val.(schema.Result)
	if !ok {
		panic("Bad foreign key value")
	}
	return fkVal
}
