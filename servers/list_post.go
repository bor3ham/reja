package servers

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
	"io/ioutil"
	"net/http"
	"strings"
)

func listPOST(
	w http.ResponseWriter,
	r *http.Request,
	c schema.Context,
	m *schema.Model,
	queryStrings map[string][]string,
	include *schema.Include,
) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	// parse the user input into instance data struct
	instance := m.Manager.Create()
	dataBlob := struct {
		Data interface{} `json:"data"`
	}{
		Data: instance,
	}
	err = JSONUnmarshal(body, &dataBlob)
	if err != nil {
		BadRequest(w, "Unable to Parse JSON", err.Error())
		return
	}

	// user cannot choose their own id
	if len(instance.GetID()) != 0 {
		BadRequest(w, "Bad Object Value", "ID's are assigned not chosen.")
		return
	}
	// type cannot be messed with
	instanceType := instance.GetType()
	if !(len(instanceType) == 0 || instanceType == m.Type) {
		BadRequest(w, "Bad Object Value", "Type does not match endpoint model.")
		return
	}

	// load defaults and validate values
	mapValues := instance.GetValues()
	values := valuesFromMap(mapValues, m.Attributes, m.Relationships)
	valueIndex := 0
	for _, attribute := range m.Attributes {
		values[valueIndex], err = attribute.DefaultFallback(values[valueIndex], instance)
		if err != nil {
			BadRequest(w, "Bad Attribute Value", err.Error())
			return
		}
		// nil values are not included in the insert statement (use db default)
		if values[valueIndex] != nil {
			values[valueIndex], err = attribute.Validate(values[valueIndex])
			if err != nil {
				BadRequest(w, "Bad Attribute Value", err.Error())
				return
			}
		}
		valueIndex += 1
	}
	for _, relation := range m.Relationships {
		values[valueIndex], err = relation.DefaultFallback(c, values[valueIndex], instance)
		if err != nil {
			BadRequest(w, "Bad Relationship Value", err.Error())
			return
		}
		// nil values are ignored
		if values[valueIndex] != nil {
			values[valueIndex], err = relation.Validate(c, values[valueIndex])
			if err != nil {
				BadRequest(w, "Bad Relationship Value", err.Error())
				return
			}
		}
		valueIndex += 1
	}

	// build insert query
	var insertColumns []string
	var insertValues []interface{}

	valueIndex = 0
	for _, attribute := range m.Attributes {
		// skip nil values (use database default)
		value := values[valueIndex]
		if value != nil {
			insertColumns = append(insertColumns, attribute.GetInsertColumns(value)...)
			insertValues = append(insertValues, attribute.GetInsertValues(value)...)
		}
		valueIndex += 1
	}
	for _, relationship := range m.Relationships {
		// skip nil values (use database default)
		value := values[valueIndex]
		if value != nil {
			insertColumns = append(insertColumns, relationship.GetInsertColumns(value)...)
			insertValues = append(insertValues, relationship.GetInsertValues(value)...)
		}
		valueIndex += 1
	}

	var valuePlaces []string
	for index, _ := range insertValues {
		valuePlaces = append(valuePlaces, fmt.Sprintf("$%d", index+1))
	}
	var query string
	if len(insertColumns) > 0 {
		query = fmt.Sprintf(
			`insert into %s (%s) values (%s) returning %s;`,
			m.Table,
			strings.Join(insertColumns, ", "),
			strings.Join(valuePlaces, ", "),
			m.IDColumn,
		)
	} else {
		query = fmt.Sprintf(
			`insert into %s default values returning %s;`,
			m.Table,
			m.IDColumn,
		)
	}

	// start a transaction
	tx, err := c.Begin()
	if err != nil {
		panic(err)
	}

	// execute insert query
	var newId string
	err = tx.QueryRow(query, insertValues...).Scan(&newId)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	// build additional queries
	var queries []schema.Query
	valueIndex = 0
	valueIndex += len(m.Attributes)
	for _, relationship := range m.Relationships {
		if values[valueIndex] != nil {
			queries = append(queries, relationship.GetInsertQueries(newId, values[valueIndex])...)
		}
		valueIndex += 1
	}

	// execute additional queries
	for _, query := range queries {
		_, err := tx.Exec(query.Query, query.Args...)
		if err != nil {
			tx.Rollback()
			panic(err)
		}
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		panic(err)
	}

	// return created object as though it were a GET
	detailGET(w, r, c, m, newId, include)
}
