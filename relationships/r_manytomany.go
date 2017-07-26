package relationships

import (
	"errors"
	"fmt"
	"github.com/bor3ham/reja/schema"
	"github.com/bor3ham/reja/utils"
	"strings"
)

type ManyToMany struct {
	RelationshipStub
	Key           string
	Table         string
	OwnIDColumn   string
	OtherIDColumn string
	OtherType     string
	Default       func(schema.Context, interface{}) PointerSet
}

func (m2m ManyToMany) GetKey() string {
	return m2m.Key
}
func (m2m ManyToMany) GetType() string {
	return m2m.OtherType
}

func (m2m ManyToMany) GetDefaultValue() interface{} {
	return schema.Page{}
}
func (m2m ManyToMany) GetValues(
	c schema.Context,
	m *schema.Model,
	ids []string,
	extra [][]interface{},
	allRelations bool,
) (
	map[string]interface{},
	map[string]map[string][]string,
) {
	if len(ids) == 0 {
		return map[string]interface{}{}, map[string]map[string][]string{}
	}
	filter := fmt.Sprintf("%s in (%s)", m2m.OwnIDColumn, strings.Join(ids, ", "))
	query := fmt.Sprintf(
		`
			select
				%s,
				%s
			from %s
			where %s
	    `,
		m2m.OwnIDColumn,
		m2m.OtherIDColumn,
		m2m.Table,
		filter,
	)
	rows, err := c.Query(query)
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
			Data:     []interface{}{},
		}
		value.Metadata["total"] = 0
		value.Metadata["count"] = 0
		values[id] = value
	}
	// go through result data
	server := c.GetServer()
	pageSize := -1
	if !allRelations {
		pageSize = server.GetIndirectPageSize()
	}
	for rows.Next() {
		var myID, otherID string
		rows.Scan(&myID, &otherID)
		value, exists := values[myID]
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

		_, exists = maps[myID]
		if !exists {
			maps[myID] = map[string][]string{}
			maps[myID][m2m.OtherType] = []string{}
		}
		maps[myID][m2m.OtherType] = append(maps[myID][m2m.OtherType], otherID)

		total += 1
		if pageSize < 0 || total <= pageSize {
			count += 1
			value.Data = append(value.Data, schema.InstancePointer{
				ID:   &otherID,
				Type: m2m.OtherType,
			})
			value.Metadata["count"] = count
		}
		value.Metadata["total"] = total
		// update the value
		values[myID] = value
	}
	// create the links
	for id, value := range values {
		total, ok := value.Metadata["total"].(int)
		if !ok {
			panic("Bad total received")
		}
		value.Links = utils.GetPaginationLinks(
			relationLink(c, m.Type, id, m2m.Key),
			1,
			pageSize,
			server.GetDefaultDirectPageSize(),
			total,
			map[string][]string{},
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

func (m2m *ManyToMany) DefaultFallback(
	c schema.Context,
	val interface{},
	instance interface{},
) (
	interface{},
	error,
) {
	var m2mVal PointerSet
	if val == nil {
		m2mVal = PointerSet{Provided: false}
	} else {
		var err error
		m2mVal, err = ParsePagePointerSet(val)
		if err != nil {
			return nil, err
		}
	}

	if !m2mVal.Provided {
		if m2m.Default != nil {
			return m2m.Default(c, instance), nil
		}
		return nil, nil
	}
	return m2mVal, nil
}
func (m2m *ManyToMany) Validate(c schema.Context, val interface{}) (interface{}, error) {
	m2mVal := AssertPointerSet(val)

	// validate the types are correct
	for _, pointer := range m2mVal.Data {
		if pointer.Type != m2m.OtherType {
			return nil, errors.New(fmt.Sprintf(
				"Relationship '%s' invalid: Incorrect type in set.",
				m2m.Key,
			))
		}
	}
	// find duplicates
	ids := map[string]bool{}
	for _, pointer := range m2mVal.Data {
		_, exists := ids[*pointer.ID]
		if exists {
			return nil, errors.New(fmt.Sprintf(
				"Relationship '%s' invalid: Duplicate object in set.",
				m2m.Key,
			))
		}
		ids[*pointer.ID] = true
	}
	// extract ids
	var instanceIds []string
	for _, pointer := range m2mVal.Data {
		instanceIds = append(instanceIds, *pointer.ID)
	}

	// check that the objects exist
	model := c.GetServer().GetModel(m2m.OtherType)
	include := schema.Include{
		Children: map[string]*schema.Include{},
	}
	instances, _, err := c.GetObjectsByIDs(model, instanceIds, &include)
	if err != nil {
		panic(err)
	}
	if len(instances) < len(ids) {
		return nil, errors.New(fmt.Sprintf(
			"Relationship '%s' invalid: Not all objects in set exist",
			m2m.Key,
		))
	}
	return m2mVal, nil
}

func (m2m *ManyToMany) GetInsertQueries(newId string, val interface{}) []schema.Query {
	m2mVal, ok := val.(PointerSet)
	if !ok {
		panic("Bad pointer set value")
	}

	var queries []schema.Query
	for _, pointer := range m2mVal.Data {
		queries = append(queries, schema.Query{
			Query: fmt.Sprintf(
				`insert into %s (%s, %s) values ($1, $2);`,
				m2m.Table,
				m2m.OwnIDColumn,
				m2m.OtherIDColumn,
			),
			Args: []interface{}{
				newId,
				*pointer.ID,
			},
		})
	}
	return queries
}

func AssertManyToMany(val interface{}) schema.Page {
	m2mVal, ok := val.(schema.Page)
	if !ok {
		panic("Bad many to many value")
	}
	return m2mVal
}
