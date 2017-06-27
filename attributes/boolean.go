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
func (b *Bool) ValidateNew(val interface{}) (interface{}, error) {
	return nil, nil
}

func AssertBool(val interface{}) *bool {
	bVal, ok := val.(**bool)
	if !ok {
		panic("Bad boolean value")
	}
	return *bVal
}
