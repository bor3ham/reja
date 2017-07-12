package relationships

import (
	"errors"
	"fmt"
	"github.com/bor3ham/reja/schema"
	"github.com/bor3ham/reja/utils"
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
	Default       func(schema.Context, interface{}) PointerSet
}

func (gfkr GenericForeignKeyReverse) GetKey() string {
	return gfkr.Key
}
func (gfkr GenericForeignKeyReverse) GetType() string {
	return gfkr.OtherType
}

func (gfkr GenericForeignKeyReverse) GetDefaultValue() interface{} {
	return schema.Page{}
}
func (gfkr GenericForeignKeyReverse) GetValues(
	c schema.Context,
	m *schema.Model,
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
	values := map[string]schema.Page{}
	maps := map[string]map[string][]string{}
	// fill in initial page data
	for _, id := range ids {
		value := schema.Page{
			Metadata: map[string]interface{}{},
			Data: []interface{}{},
		}
		value.Metadata["total"] = 0
		value.Metadata["count"] = 0
		values[id] = value
	}
	// go through result data
	server := c.GetServer()
	pageSize := server.GetIndirectPageSize()
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
			value.Data = append(value.Data, schema.InstancePointer{
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
	// create the links
	for id, value := range values {
		total, ok := value.Metadata["total"].(int)
		if !ok {
			panic("Bad total received")
		}
		value.Links = utils.GetPaginationLinks(
			relationLink(c, m.Type, id, gfkr.Key),
			1,
			pageSize,
			server.GetDefaultDirectPageSize(),
			total,
			map[string]string{},
		)
		values[id] = value
	}
	// generalise values
	generalValues := map[string]interface{}{}
	for id, value := range values {
		generalValues[id] = value
	}
	return generalValues, maps
}

func (gfkr *GenericForeignKeyReverse) DefaultFallback(
	c schema.Context,
	val interface{},
	instance interface{},
) (
	interface{},
	error,
) {
	gfkrVal, err := ParsePagePointerSet(val)
	if err != nil {
		return nil, err
	}
	if !gfkrVal.Provided {
		if gfkr.Default != nil {
			return gfkr.Default(c, instance), nil
		}
		return nil, nil
	}
	return gfkrVal, nil
}
func (gfkr *GenericForeignKeyReverse) Validate(
	c schema.Context,
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
	include := schema.Include{
		Children: map[string]*schema.Include{},
	}
	instances, _, err := c.GetObjects(
		model,
		instanceIds,
		0,
		0,
		&include,
	)
	if err != nil {
		panic(err)
	}
	if len(instances) < len(ids) {
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
) []schema.Query {
	gfkrVal, ok := val.(PointerSet)
	if !ok {
		panic("Bad pointer set value")
	}

	var ids []string
	for _, pointer := range gfkrVal.Data {
		ids = append(ids, *pointer.ID)
	}

	if len(ids) == 0 {
		return []schema.Query{}
	}

	query := fmt.Sprintf(
		`update %s set (%s, %s) = ($1, $2) where %s in (%s);`,
		gfkr.Table,
		gfkr.OwnTypeColumn,
		gfkr.OwnIDColumn,
		gfkr.OtherIDColumn,
		strings.Join(ids, ", "),
	)
	return []schema.Query{
		schema.Query{
			Query: query,
			Args: []interface{}{
				gfkr.OwnType,
				newId,
			},
		},
	}
}

func AssertGenericForeignKeyReverse(val interface{}) schema.Page {
	gfkrVal, ok := val.(schema.Page)
	if !ok {
		panic("Bad generic foreign key reverse value")
	}
	return gfkrVal
}
