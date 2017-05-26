package attributes

type Text struct {
  ColumnName string
}

func (a Text) GetColumns() []string {
  return []string{a.ColumnName}
}
