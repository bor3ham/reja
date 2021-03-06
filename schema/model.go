package schema

import (
	"errors"
	"fmt"
	"strings"
)

type Model struct {
	Type          string
	Table         string
	IDColumn      string
	IDGenerator   func(Context) string
	DefaultOrder  string
	Attributes    []Attribute
	Relationships []Relationship
	Manager       Manager
}

func (m Model) DirectFields() ([]string, []interface{}) {
	var allColumns []string
	var allVars []interface{}
	for _, attribute := range m.Attributes {
		columns, vars := attribute.GetSelectDirect()
		allColumns = append(allColumns, columns...)
		allVars = append(allVars, vars...)
	}
	return allColumns, allVars
}

func (m Model) ExtraFields() ([]string, [][]interface{}) {
	var allColumns []string
	var allVars [][]interface{}
	for _, relationship := range m.Relationships {
		columns, vars := relationship.GetSelectExtra()
		allColumns = append(allColumns, columns...)
		allVars = append(allVars, vars)
	}
	return allColumns, allVars
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
