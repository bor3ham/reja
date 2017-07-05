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

type GenericForeignKeyReverse struct {
	RelationshipStub
	Key           string
	Table         string
	OwnType       string
	OwnTypeColumn string
	OwnIDColumn   string
	OtherIDColumn string
	OtherType     string
	Default       func(context.Context, interface{}) PointerSet
}

func (gfkr GenericForeignKeyReverse) GetKey() string {
	return gfkr.Key
}
func (gfkr GenericForeignKeyReverse) GetType() string {
	return gfkr.OtherType
}

func (gfkr GenericForeignKeyReverse) GetDefaultValue() interface{} {
	return format.Page{}
}
func (gfkr GenericForeignKeyReverse) GetValues(
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

	idFilter := fmt.Sprintf("%s in (%s)", gfkr.OwnIDColumn, strings.Join(ids, ", "))
	typeFilter := fmt.Sprintf("%s = $1", gfkr.OwnTypeColumn)
	query := fmt.Sprintf(
		`
			select
				%s,
				%s
			from %s
			where (%s and %s)
	    `,
		gfkr.OtherIDColumn,
		gfkr.OwnIDColumn,
		gfkr.Table,
		idFilter,
		typeFilter,
	)
	rows, err := c.Query(query, gfkr.OwnType)
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
	pageSize := c.GetServer().GetIndirectPageSize()
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
		if total <= pageSize {
			count += 1
			value.Data = append(value.Data, instances.InstancePointer{
				ID:   &otherId,
				Type: gfkr.OtherType,
			})
			value.Metadata["count"] = count
		}
		value.Metadata["total"] = total
		// update the value
		values[ownId] = value

		// add to maps
		_, exists = maps[ownId]
		if !exists {
			maps[ownId] = map[string][]string{}
			maps[ownId][gfkr.OtherType] = []string{}
		}
		maps[ownId][gfkr.OtherType] = append(maps[ownId][gfkr.OtherType], otherId)
	}
	// generalise values
	generalValues := map[string]interface{}{}
	for id, value := range values {
		generalValues[id] = value
	}
	return generalValues, maps
}

func (gfkr *GenericForeignKeyReverse) DefaultFallback(
	c context.Context,
	val interface{},
	instance interface{},
) interface{} {
	gfkrVal, err := ParsePagePointerSet(val)
	if err != nil {
		panic(err)
	}
	if !gfkrVal.Provided {
		if gfkr.Default != nil {
			return gfkr.Default(c, instance)
		}
		return nil
	}
	return gfkrVal
}
func (gfkr *GenericForeignKeyReverse) Validate(
	c context.Context,
	val interface{},
) (
	interface{},
	error,
) {
	gfkrVal := AssertPointerSet(val)

	// validate the types are correct
	for _, pointer := range gfkrVal.Data {
		if pointer.Type != gfkr.OtherType {
			return nil, errors.New(fmt.Sprintf(
				"Relationship '%s' invalid: Incorrect type in set.",
				gfkr.Key,
			))
		}
	}
	// find duplicates
	ids := map[string]bool{}
	for _, pointer := range gfkrVal.Data {
		_, exists := ids[*pointer.ID]
		if exists {
			return nil, errors.New(fmt.Sprintf(
				"Relationship '%s' invalid: Duplicate object in set.",
				gfkr.Key,
			))
		}
		ids[*pointer.ID] = true
	}
	// extract ids
	var instanceIds []string
	for _, pointer := range gfkrVal.Data {
		instanceIds = append(instanceIds, *pointer.ID)
	}

	// check that the objects exist
	model := c.GetServer().GetModel(gfkr.OtherType)
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
			gfkr.Key,
		))
	}
	return gfkrVal, nil
}

func (gfkr *GenericForeignKeyReverse) GetInsertQueries(
	newId string,
	val interface{},
) []database.QueryBlob {
	gfkrVal, ok := val.(PointerSet)
	if !ok {
		panic("Bad pointer set value")
	}

	var ids []string
	for _, pointer := range gfkrVal.Data {
		ids = append(ids, *pointer.ID)
	}

	if len(ids) == 0 {
		return []database.QueryBlob{}
	}

	query := fmt.Sprintf(
		`update %s set (%s, %s) = ($1, $2) where %s in (%s);`,
		gfkr.Table,
		gfkr.OwnTypeColumn,
		gfkr.OwnIDColumn,
		gfkr.OtherIDColumn,
		strings.Join(ids, ", "),
	)
	return []database.QueryBlob{
		database.QueryBlob{
			Query: query,
			Args: []interface{}{
				gfkr.OwnType,
				newId,
			},
		},
	}
}

func AssertGenericForeignKeyReverse(val interface{}) format.Page {
	gfkrVal, ok := val.(format.Page)
	if !ok {
		panic("Bad generic foreign key reverse value")
	}
	return gfkrVal
}
