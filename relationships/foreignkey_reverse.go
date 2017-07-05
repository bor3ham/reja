package relationships

import (
	"errors"
	"fmt"
	"github.com/bor3ham/reja/context"
	"github.com/bor3ham/reja/database"
	"github.com/bor3ham/reja/format"
	"github.com/bor3ham/reja/instances"
	"github.com/bor3ham/reja/models"
	"strings"
)

const defaultPageSize = 5

type ForeignKeyReverse struct {
	RelationshipStub
	Key            string
	SourceTable    string
	SourceIDColumn string
	ColumnName     string
	Type           string
	Default        func(interface{}) PointerSet
}

func (fkr ForeignKeyReverse) GetKey() string {
	return fkr.Key
}
func (fkr ForeignKeyReverse) GetType() string {
	return fkr.Type
}

func (fkr ForeignKeyReverse) GetDefaultValue() interface{} {
	return format.Page{}
}
func (fkr ForeignKeyReverse) GetValues(
	c context.Context,
	ids []string,
	extra [][]interface{},
) (
	map[string]interface{},
	map[string]map[string][]string,
) {
	if len(ids) == 0 {
		return map[string]interface{}{}, map[string]map[string][]string{}
	}
	filter := fmt.Sprintf("%s in (%s)", fkr.ColumnName, strings.Join(ids, ", "))

	query := fmt.Sprintf(
		`
      select
        %s,
        %s
      from %s
      where %s
    `,
		fkr.SourceIDColumn,
		fkr.ColumnName,
		fkr.SourceTable,
		filter,
	)
	rows, err := c.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	values := map[string]format.Page{}
	maps := map[string]map[string][]string{}
	// fill in initial page data
	for _, id := range ids {
		value := format.Page{
			Metadata: map[string]interface{}{},
			Links:    map[string]*string{},
			Data:     []interface{}{},
		}
		value.Metadata["total"] = 0
		value.Metadata["count"] = 0
		values[id] = value
	}
	// go through result data
	for rows.Next() {
		var otherId, ownId string
		rows.Scan(&otherId, &ownId)
		value, exists := values[ownId]
		if !exists {
			panic("Found unexpected id in results")
		}
		total, ok := value.Metadata["total"].(int)
		if !ok {
			panic("Bad total received")
		}
		count, ok := value.Metadata["count"].(int)
		if !ok {
			panic("Bad count received")
		}
		total += 1
		if total <= defaultPageSize {
			count += 1
			value.Data = append(value.Data, instances.InstancePointer{
				ID:   &otherId,
				Type: fkr.Type,
			})
			value.Metadata["count"] = count
		}
		value.Metadata["total"] = total

		// add to maps
		_, exists = maps[ownId]
		if !exists {
			maps[ownId] = map[string][]string{}
			maps[ownId][fkr.Type] = []string{}
		}
		maps[ownId][fkr.Type] = append(maps[ownId][fkr.Type], otherId)
		// update value
		values[ownId] = value
	}
	// generalise values
	generalValues := map[string]interface{}{}
	for id, value := range values {
		generalValues[id] = value
	}
	return generalValues, maps
}

func (fkr *ForeignKeyReverse) DefaultFallback(val interface{}, instance interface{}) interface{} {
	fkrVal, err := ParsePagePointerSet(val)
	if err != nil {
		panic(err)
	}
	if !fkrVal.Provided {
		if fkr.Default != nil {
			return fkr.Default(instance)
		}
		return nil
	}
	return fkrVal
}
func (fkr *ForeignKeyReverse) Validate(c context.Context, val interface{}) (interface{}, error) {
	fkrVal := AssertPointerSet(val)

	// validate the types are correct
	for _, pointer := range fkrVal.Data {
		if pointer.Type != fkr.Type {
			return nil, errors.New(fmt.Sprintf(
				"Relationship '%s' invalid: Incorrect type in set.",
				fkr.Key,
			))
		}
	}
	// find duplicates
	ids := map[string]bool{}
	for _, pointer := range fkrVal.Data {
		_, exists := ids[*pointer.ID]
		if exists {
			return nil, errors.New(fmt.Sprintf(
				"Relationship '%s' invalid: Duplicate object in set.",
				fkr.Key,
			))
		}
		ids[*pointer.ID] = true
	}
	// extract ids
	var instanceIds []string
	for _, pointer := range fkrVal.Data {
		instanceIds = append(instanceIds, *pointer.ID)
	}

	// check that the objects exist
	model := models.GetModel(fkr.Type)
	include := models.Include{
		Children: map[string]*models.Include{},
	}
	instances, _, err := models.GetObjects(
		c,
		*model,
		instanceIds,
		0,
		0,
		&include,
	)
	if err != nil {
		panic(err)
	}
	if len(instances) != len(ids) {
		return nil, errors.New(fmt.Sprintf(
			"Relationship '%s' invalid: Not all objects in set exist",
			fkr.Key,
		))
	}
	return fkrVal, nil
}

func (fkr *ForeignKeyReverse) GetInsertQueries(newId string, val interface{}) []database.QueryBlob {
	fkrVal, ok := val.(PointerSet)
	if !ok {
		panic("Bad pointer set value")
	}

	var ids []string
	for _, pointer := range fkrVal.Data {
		ids = append(ids, *pointer.ID)
	}

	if len(ids) == 0 {
		return []database.QueryBlob{}
	}

	query := fmt.Sprintf(
		`update %s set %s = $1 where %s in (%s);`,
		fkr.SourceTable,
		fkr.ColumnName,
		fkr.SourceIDColumn,
		strings.Join(ids, ", "),
	)
	return []database.QueryBlob{
		database.QueryBlob{
			Query: query,
			Args: []interface{}{
				newId,
			},
		},
	}
}

func AssertForeignKeyReverse(val interface{}) format.Page {
	fkrVal, ok := val.(format.Page)
	if !ok {
		panic("Bad foreign key reverse value")
	}
	return fkrVal
}
