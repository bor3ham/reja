package relationships

import (
  "fmt"
  "strings"
  "github.com/bor3ham/reja/database"
)

type ForeignKeyReverse struct {
  SourceTable string
  SourceIDColumn string
  ColumnName string
  Type string
}

func (fkr ForeignKeyReverse) GetColumnNames() []string {
  return []string{}
}
func (fkr ForeignKeyReverse) GetColumnVariables() []interface{} {
  return []interface{}{}
}

func (fkr ForeignKeyReverse) GetDefaultValue() interface{} {
  return &Pointers{
    Data: []*PointerData{},
  }
}
func (fkr ForeignKeyReverse) GetValues(ids []string) map[string]interface{} {
  filter := fmt.Sprintf("%s in (%s)", fkr.ColumnName, strings.Join(ids, ", "))

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
    fkr.SourceIDColumn,
    fkr.ColumnName,
    fkr.SourceTable,
    filter,
  )
  rows, err := database.Query(query)
  if err != nil {
    panic(err)
  }
  defer rows.Close()
  values := map[string]*Pointers{}
  for rows.Next() {
    var id, my_id string
    rows.Scan(&id, &my_id)
    value, exists := values[my_id]
    if !exists {
      value = &Pointers{}
      values[my_id] = value
    }
    value.Data = append(value.Data, &PointerData{
      ID: &id,
      Type: fkr.Type,
    })
  }
  general_values := map[string]interface{}{}
  for id, value := range values {
    general_values[id] = value
  }
  return general_values
}
