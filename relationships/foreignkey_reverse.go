package relationships

import (
	"fmt"
	"github.com/bor3ham/reja/context"
	"github.com/bor3ham/reja/format"
	"strings"
)

const defaultPageSize = 5

type ForeignKeyReverse struct {
	Key            string
	SourceTable    string
	SourceIDColumn string
	ColumnName     string
	Type           string
}

func (fkr ForeignKeyReverse) GetKey() string {
	return fkr.Key
}
func (fkr ForeignKeyReverse) GetType() string {
	return fkr.Type
}

func (fkr ForeignKeyReverse) GetInstanceColumnNames() []string {
	return []string{}
}
func (fkr ForeignKeyReverse) GetInstanceColumnVariables() []interface{} {
	return []interface{}{}
}
func (fkr ForeignKeyReverse) GetExtraColumnNames() []string {
	return []string{}
}
func (fkr ForeignKeyReverse) GetExtraColumnVariables() []interface{} {
	return []interface{}{}
}

func (fkr ForeignKeyReverse) GetDefaultValue() interface{} {
	return &Pointers{
		Data: []*PointerData{},
	}
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
	}
	// generalise values
	generalValues := map[string]interface{}{}
	for id, value := range values {
		generalValues[id] = value
	}
	return generalValues, maps
}

func (fkr *ForeignKeyReverse) ValidateNew(val interface{}) (interface{}, error) {
	return nil, nil
}

func AssertForeignKeyReverse(val interface{}) *format.Page {
	fkrVal, ok := val.(*format.Page)
	if !ok {
		panic("Bad foreign key reverse value")
	}
	return fkrVal
}
