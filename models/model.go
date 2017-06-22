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

func (m Model) FieldVariables() []interface{} {
	var fields []interface{}
	for _, attribute := range m.Attributes {
		fields = append(fields, attribute.GetColumnVariables()...)
	}
	for _, relationship := range m.Relationships {
		fields = append(fields, relationship.GetInstanceColumnVariables()...)
	}
	return fields
}
func (m Model) FieldNames() []string {
	var columns []string
	for _, attribute := range m.Attributes {
		columns = append(columns, attribute.GetColumnNames()...)
	}
	for _, relationship := range m.Relationships {
		columns = append(columns, relationship.GetInstanceColumnNames()...)
	}
	return columns
}

func (m Model) ExtraVariables() [][]interface{} {
	var fields [][]interface{}
	for _, relationship := range m.Relationships {
		fields = append(fields, relationship.GetExtraColumnVariables())
	}
	return fields
}
func (m Model) ExtraNames() []string {
	var columns []string
	for _, relationship := range m.Relationships {
		columns = append(columns, relationship.GetExtraColumnNames()...)
	}
	return columns
}

func logQueryCount(count int) {
	fmt.Println("Database queries:", count)
}
