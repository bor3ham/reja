package relationships

type ForeignKey struct {
  ColumnName string
  Type string
}

func (a ForeignKey) GetColumns() []string {
  return []string{a.ColumnName}
}

func (a ForeignKey) GetEmptyKeyedValue() interface{} {
  return nil
}

func (a ForeignKey) GetKeyedValues(ids []string) map[int]interface{} {
  return map[int]interface{}{}
}
