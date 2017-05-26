package relationships

type ForeignKey struct {
  ColumnName string
  Type string
}

func (a ForeignKey) GetColumns() []string {
  return []string{a.ColumnName}
}
