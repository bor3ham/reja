package attributes

type Bool struct {
	ColumnName string
}

func (b Bool) GetColumnNames() []string {
	return []string{b.ColumnName}
}
func (b Bool) GetColumnVariables() []interface{} {
	var destination *bool
	return []interface{}{
		&destination,
	}
}
