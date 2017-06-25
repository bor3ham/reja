package relationships

import (
	"fmt"
	"github.com/bor3ham/reja/context"
	"github.com/bor3ham/reja/format"
	"strings"
)

type GenericForeignKeyReverse struct {
	Key            string
	Table    string
	OwnType string
	OwnTypeColumn string
	OwnIDColumn string
	OtherIDColumn string
	OtherType           string
}

func (gfkr GenericForeignKeyReverse) GetKey() string {
	return gfkr.Key
}
func (gfkr GenericForeignKeyReverse) GetType() string {
	return gfkr.OtherType
}

func (gfkr GenericForeignKeyReverse) GetInstanceColumnNames() []string {
	return []string{}
}
func (gfkr GenericForeignKeyReverse) GetInstanceColumnVariables() []interface{} {
	return []interface{}{}
}
func (gfkr GenericForeignKeyReverse) GetExtraColumnNames() []string {
	return []string{}
}
func (gfkr GenericForeignKeyReverse) GetExtraColumnVariables() []interface{} {
	return []interface{}{}
}

func (gfkr GenericForeignKeyReverse) GetDefaultValue() interface{} {
	return &Pointers{
		Data: []*PointerData{},
	}
}
func (gfkr GenericForeignKeyReverse) GetValues(
	c context.Context,
	ids []string,
	extra [][]interface{},
) (
	map[string]interface{},
	map[string][]string,
) {
	if len(ids) == 0 {
		return map[string]interface{}{}, map[string][]string{}
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
			value.Data = append(value.Data, PointerData{
				ID:   &otherId,
				Type: gfkr.OtherType,
			})
			value.Metadata["count"] = count
		}
		value.Metadata["total"] = total

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
	return generalValues, flattenMaps(maps)
}
