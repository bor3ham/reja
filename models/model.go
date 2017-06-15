package models

import (
	"fmt"
	"github.com/bor3ham/reja/attributes"
	"github.com/bor3ham/reja/managers"
	"github.com/bor3ham/reja/relationships"
)

type Model struct {
	Type          string
	Table         string
	IDColumn      string
	Attributes    []attributes.Attribute
	Relationships []relationships.Relationship
	Manager       managers.Manager
}

type RelationResult struct {
	Values  map[string]interface{}
	Default interface{}
}

func (m Model) FieldVariables() []interface{} {
	var fields []interface{}
	for _, attribute := range m.Attributes {
		fields = append(fields, attribute.GetColumnVariables()...)
	}
	for _, relationship := range m.Relationships {
		fields = append(fields, relationship.GetColumnVariables()...)
	}
	return fields
}

func (m Model) FieldNames() []string {
	var columns []string
	for _, attribute := range m.Attributes {
		columns = append(columns, attribute.GetColumnNames()...)
	}
	for _, relationship := range m.Relationships {
		columns = append(columns, relationship.GetColumnNames()...)
	}
	return columns
}

func logQueryCount(count int) {
	fmt.Println("Database queries:", count)
}
