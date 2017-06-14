package attributes

type Text struct {
	ColumnName string
}

func (t Text) GetColumnNames() []string {
	return []string{t.ColumnName}
}
func (t Text) GetColumnVariables() []interface{} {
	var destination *string
	return []interface{}{
		&destination,
	}
}
