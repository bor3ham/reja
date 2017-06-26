package attributes

type Integer struct {
	ColumnName string
}

func (i Integer) GetColumnNames() []string {
	return []string{i.ColumnName}
}
func (i Integer) GetColumnVariables() []interface{} {
	var destination *int
	return []interface{}{
		&destination,
	}
}

func AssertInteger(val interface{}) *int {
	intVal, ok := val.(**int)
	if !ok {
		panic("Bad integer value")
	}
	return *intVal
}
