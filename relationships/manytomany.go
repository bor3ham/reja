package relationships

import (
	"fmt"
	"github.com/bor3ham/reja/context"
	"strings"
)

type ManyToMany struct {
	Table    string
	OwnIDColumn string
	OtherIDColumn string
	OtherType string
}

func (m2m ManyToMany) GetColumnNames() []string {
	return []string{}
}
func (m2m ManyToMany) GetColumnVariables() []interface{} {
	return []interface{}{}
}

func (m2m ManyToMany) GetDefaultValue() interface{} {
	return &Pointers{
		Data: []*PointerData{},
	}
}
func (m2m ManyToMany) GetValues(c context.Context, ids []string) map[string]interface{} {
	if len(ids) == 0 {
		return map[string]interface{}{}
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
	values := map[string]*Pointers{}
	for rows.Next() {
		var myID, otherID string
		rows.Scan(&myID, &otherID)
		value, exists := values[myID]
		if !exists {
			value = &Pointers{}
			values[myID] = value
		}
		value.Data = append(value.Data, &PointerData{
			ID:   &otherID,
			Type: m2m.OtherType,
		})
	}
	generalValues := map[string]interface{}{}
	for id, value := range values {
		generalValues[id] = value
	}
	return generalValues
}
