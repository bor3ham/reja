package schema

import (
	"fmt"
	"strings"
	"errors"
)

type Model struct {
	Type          string
	Table         string
	IDColumn      string
	DefaultOrder  string
	Attributes    []Attribute
	Relationships []Relationship
	Manager       Manager
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

func (m Model) GetOrderQuery(asParam string) (string, string, error) {
	validParam := ""

	validOrders := map[string]string{
		"id": m.IDColumn,
	}
	for _, attribute := range m.Attributes {
		attrOrders := attribute.GetOrderMap()
		for key, arg := range attrOrders {
			validOrders[key] = arg
		}
	}

	queryArgs := []string{}
	splitOrders := strings.Split(asParam, ",")
	orderedColumns := map[string]bool{}
	for _, order := range splitOrders {
		cleanOrder := strings.ToLower(strings.TrimSpace(order))
		if len(cleanOrder) == 0 {
			continue
		}
		posCleanOrder := strings.TrimPrefix(cleanOrder, "-")
		column, exists := validOrders[posCleanOrder]
		if !exists {
			return "", "", errors.New(fmt.Sprintf(
				"Cannot order by unknown field '%s'.",
				cleanOrder,
			))
		}
		_, exists = orderedColumns[column]
		if exists {
			return "", "", errors.New(fmt.Sprintf(
				"Cannot order by column '%s' twice.",
				cleanOrder,
			))
		}
		orderedColumns[column] = true
		query := column
		if posCleanOrder != cleanOrder {
			query += " desc"
		}
		queryArgs = append(queryArgs, query)
		if len(validParam) != 0 {
			validParam += ","
		}
		validParam += cleanOrder
	}

	query := ""
	if len(queryArgs) > 0 {
		query = fmt.Sprintf(
			"order by %s",
			strings.Join(queryArgs, ", "),
		)
	}

	if validParam == m.DefaultOrder {
		validParam = ""
	}

	return query, validParam, nil
}
