package relationships

import (
  "fmt"
  "github.com/bor3ham/reja/database"
)

type ForeignKeyReverse struct {
  SourceTable string
  SourceIDColumn string
  ColumnName string
  Type string
}

func (a ForeignKeyReverse) GetColumns() []string {
  return []string{}
}

func (a ForeignKeyReverse) GetEmptyKeyedValue() interface{} {
  return &Pointers{
    Data: []*PointerData{},
  }
}

// get keyed values returns dictionary values (from query)
// dictionary id: value
func (a ForeignKeyReverse) GetKeyedValues(filter string) map[int]interface{} {
  // where id = 3
  // where id in (1,2,3,4,5,6,7,8)
  query := fmt.Sprintf(
    `
      select
        %s,
        %s
      from %s
      where %s
    `,
    a.SourceIDColumn,
    a.ColumnName,
    a.SourceTable,
    filter,
  )
  rows, err := database.Query(query)
  if err != nil {
    panic(err)
  }
  defer rows.Close()
  values := map[int]*Pointers{}
  for rows.Next() {
    var id, my_id int
    rows.Scan(&id, &my_id)
    value, exists := values[my_id]
    if !exists {
      value = &Pointers{}
      values[my_id] = value
    }
    value.Data = append(value.Data, &PointerData{
      ID: id,
      Type: a.Type,
    })
  }
  // ewww
  return_values := map[int]interface{}{}
  for id, value := range values {
    return_values[id] = value
  }
  return return_values
}
