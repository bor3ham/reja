package relationships

import (
	"fmt"
	"github.com/bor3ham/reja/context"
	"github.com/bor3ham/reja/format"
	"strings"
)

type ManyToMany struct {
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

func (m2m ManyToMany) GetInstanceColumnNames() []string {
	return []string{}
}
func (m2m ManyToMany) GetInstanceColumnVariables() []interface{} {
	return []interface{}{}
}
func (m2m ManyToMany) GetExtraColumnNames() []string {
	return []string{}
}
func (m2m ManyToMany) GetExtraColumnVariables() []interface{} {
	return []interface{}{}
}

func (m2m ManyToMany) GetDefaultValue() interface{} {
	return &Pointers{
		Data: []*PointerData{},
	}
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
	values := map[string]*format.Page{}
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
		values[id] = &value
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
			value.Data = append(value.Data, PointerData{
				ID:   &otherID,
				Type: m2m.OtherType,
			})
			value.Metadata["count"] = count
		}
		value.Metadata["total"] = total
	}
	// generalise values
	generalValues := map[string]interface{}{}
	for id, value := range values {
		generalValues[id] = value
	}
	return generalValues, maps
}

func (m2m *ManyToMany) ValidateNew(val interface{}) error {
	return nil
}

func AssertManyToMany(val interface{}) *format.Page {
	m2mVal, ok := val.(*format.Page)
	if !ok {
		panic("Bad many to many value")
	}
	return m2mVal
}
