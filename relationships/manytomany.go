package relationships

import (
	"errors"
	"fmt"
	"github.com/bor3ham/reja/context"
	"github.com/bor3ham/reja/format"
	"github.com/bor3ham/reja/instances"
	"github.com/bor3ham/reja/models"
	"github.com/bor3ham/reja/database"
	"strings"
)

type ManyToMany struct {
	RelationshipStub
	Key           string
	Table         string
	OwnIDColumn   string
	OtherIDColumn string
	OtherType     string
}

func (m2m ManyToMany) GetKey() string {
	return m2m.Key
}
func (m2m ManyToMany) GetType() string {
	return m2m.OtherType
}

func (m2m ManyToMany) GetDefaultValue() interface{} {
	return format.Page{}
}
func (m2m ManyToMany) GetValues(
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
		if total <= defaultPageSize {
			count += 1
			value.Data = append(value.Data, instances.InstancePointer{
				ID:   &otherID,
				Type: m2m.OtherType,
			})
			value.Metadata["count"] = count
		}
		value.Metadata["total"] = total
		// update the value
		values[myID] = value
	}
	// generalise values
	generalValues := map[string]interface{}{}
	for id, value := range values {
		generalValues[id] = value
	}
	return generalValues, maps
}

func (m2m *ManyToMany) ValidateNew(c context.Context, val interface{}) (interface{}, error) {
	var m2mVal PointerSet
	if val == nil {
		m2mVal = PointerSet{}
	} else {
		var err error
		m2mVal, err = ParsePagePointerSet(val)
		if err != nil {
			return nil, errors.New(fmt.Sprintf(
				"Relationship '%s' invalid: %s",
				m2m.Key,
				err.Error(),
			))
		}
	}
	return m2m.validate(c, m2mVal)
}
func (m2m *ManyToMany) validate(c context.Context, val PointerSet) (interface{}, error) {
	// validate the types are correct
	for _, pointer := range val.Data {
		if pointer.Type != m2m.OtherType {
			return nil, errors.New(fmt.Sprintf(
				"Relationship '%s' invalid: Incorrect type in set.",
				m2m.Key,
			))
		}
	}
	// find duplicates
	ids := map[string]bool{}
	for _, pointer := range val.Data {
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
	for _, pointer := range val.Data {
		instanceIds = append(instanceIds, *pointer.ID)
	}

	// check that the objects exist
	model := models.GetModel(m2m.OtherType)
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
			m2m.Key,
		))
	}
	return val, nil
}

func (m2m *ManyToMany) GetInsertQueries(newId string, val interface{}) []database.QueryBlob {
	m2mVal, ok := val.(PointerSet)
	if !ok {
		panic("Bad pointer set value")
	}

	var ids []string
	for _, pointer := range m2mVal.Data {
		ids = append(ids, *pointer.ID)
	}

	var queries []database.QueryBlob
	for _, id := range ids {
		queries = append(queries, database.QueryBlob{
			Query: fmt.Sprintf(
				`insert into %s (%s, %s) values ($1, $2);`,
				m2m.Table,
				m2m.OwnIDColumn,
				m2m.OtherIDColumn,
			),
			Args: []interface{}{
				newId,
				id,
			},
		})
	}
	return queries
}

func AssertManyToMany(val interface{}) format.Page {
	m2mVal, ok := val.(format.Page)
	if !ok {
		panic("Bad many to many value")
	}
	return m2mVal
}
