package models

import (
	"fmt"
	"github.com/bor3ham/reja/attributes"
	"github.com/bor3ham/reja/database"
	"github.com/bor3ham/reja/managers"
)

type Relationship interface {
	GetKey() string
	GetType() string

	GetSelectDirectColumns() []string
	GetSelectDirectVariables() []interface{}
	GetSelectExtraColumns() []string
	GetSelectExtraVariables() []interface{}

	GetDefaultValue() interface{}
	GetValues(
		Context,
		[]string,
		[][]interface{},
	) (
		map[string]interface{},
		map[string]map[string][]string,
	)

	GetInsertQueries(string, interface{}) []database.QueryBlob

	DefaultFallback(Context, interface{}, interface{}) interface{}
	Validate(Context, interface{}) (interface{}, error)
}

type Model struct {
	Type          string
	Table         string
	IDColumn      string
	Attributes    []attributes.Attribute
	Relationships []Relationship
	Manager       managers.Manager
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

func logQueryCount(count int) {
	fmt.Println("Database queries:", count)
}
