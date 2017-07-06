package server

import (
	"github.com/bor3ham/reja/schema"
)

type Model struct {
	Type          string
	Table         string
	IDColumn      string
	Attributes    []schema.Attribute
	Relationships []schema.Relationship
	Manager       schema.Manager
}

func (m *Model) GetType() string {
	return m.Type
}
func (m *Model) GetTable() string {
	return m.Table
}
func (m *Model) GetIDColumn() string {
	return m.IDColumn
}
func (m *Model) GetRelationships() []schema.Relationship {
	return m.Relationships
}
func (m *Model) GetManager() schema.Manager {
	return m.Manager
}

func (m Model) FieldColumns() []string {
	var columns []string
	for _, attribute := range m.Attributes {
		columns = append(columns, attribute.GetSelectDirectColumns()...)
	}
	for _, relationship := range m.Relationships {
		columns = append(columns, relationship.GetSelectDirectColumns()...)
	}
	return columns
}
func (m Model) FieldVariables() []interface{} {
	var fields []interface{}
	for _, attribute := range m.Attributes {
		fields = append(fields, attribute.GetSelectDirectVariables()...)
	}
	for _, relationship := range m.Relationships {
		fields = append(fields, relationship.GetSelectDirectVariables()...)
	}
	return fields
}

func (m Model) ExtraColumns() []string {
	var columns []string
	for _, relationship := range m.Relationships {
		columns = append(columns, relationship.GetSelectExtraColumns()...)
	}
	return columns
}
func (m Model) ExtraVariables() [][]interface{} {
	var fields [][]interface{}
	for _, relationship := range m.Relationships {
		fields = append(fields, relationship.GetSelectExtraVariables())
	}
	return fields
}
