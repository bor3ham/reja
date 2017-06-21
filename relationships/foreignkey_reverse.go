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
func (fkr ForeignKeyReverse) GetValues(c context.Context, ids []string, extra [][]interface{}) (map[string]interface{}, []string) {
	if len(ids) == 0 {
		return map[string]interface{}{}, []string{}
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
	relationIds := []string{}
	for rows.Next() {
		var id, my_id string
		rows.Scan(&id, &my_id)
		relationIds = append(relationIds, id)
		value, exists := values[my_id]
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
				ID:   &id,
				Type: fkr.Type,
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
	return generalValues, relationIds
}
