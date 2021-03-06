package servers

import (
	"encoding/json"
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

	err = json.Unmarshal(body, &dataBlob)
	if err != nil {
		BadRequest(c, w, "Unable to Parse JSON", err.Error())
		return
	}

	// user cannot choose their own id
	if len(instance.GetID()) != 0 {
		BadRequest(c, w, "Bad Object Value", "ID's are assigned not chosen.")
		return
	}
	// type cannot be messed with
	instanceType := instance.GetType()
	if !(len(instanceType) == 0 || instanceType == m.Type) {
		BadRequest(c, w, "Bad Object Value", "Type does not match endpoint model.")
		return
	}

	// load defaults and validate values
	mapValues := instance.GetValues()
	values := valuesFromMap(mapValues, m.Attributes, m.Relationships)
	valueIndex := 0
	for _, attribute := range m.Attributes {
		values[valueIndex], err = attribute.DefaultFallback(values[valueIndex], instance)
		if err != nil {
			BadRequest(c, w, "Bad Attribute Value", err.Error())
			return
		}
		// nil values are not included in the insert statement (use db default)
		if values[valueIndex] != nil {
			values[valueIndex], err = attribute.Validate(values[valueIndex])
			if err != nil {
				BadRequest(c, w, "Bad Attribute Value", err.Error())
				return
			}
		}
		valueIndex += 1
	}
	for _, relation := range m.Relationships {
		values[valueIndex], err = relation.DefaultFallback(c, values[valueIndex], instance)
		if err != nil {
			BadRequest(c, w, "Bad Relationship Value", err.Error())
			return
		}
		// nil values are ignored
		if values[valueIndex] != nil {
			values[valueIndex], err = relation.Validate(c, values[valueIndex])
			if err != nil {
				BadRequest(c, w, "Bad Relationship Value", err.Error())
				return
			}
		}
		valueIndex += 1
	}

	// run manager validation
	mapValues = mapFromValues(values, m.Attributes, m.Relationships)
	// instance = m.Manager.Create()
	// instance.SetValues(mapValues)
	err = m.Manager.BeforeCreate(c, mapValues)
	if err != nil {
		BadRequest(c, w, "Bad New Instance", err.Error())
		return
	}
	err = m.Manager.BeforeSave(c, mapValues)
	if err != nil {
		BadRequest(c, w, "Bad Instance", err.Error())
		return
	}

	// build insert query
	var insertColumns []string
	var insertValues []interface{}

	valueIndex = 0
	// get id if determined
	if m.IDGenerator != nil {
		insertColumns = append(insertColumns, m.IDColumn)
		insertValues = append(insertValues, m.IDGenerator(c))
	}
	for _, attribute := range m.Attributes {
		// skip nil values (use database default)
		value := values[valueIndex]
		if value != nil {
			columns, values := attribute.GetInsert(value)
			insertColumns = append(insertColumns, columns...)
			insertValues = append(insertValues, values...)
		}
		valueIndex += 1
	}
	for _, relationship := range m.Relationships {
		// skip nil values (use database default)
		value := values[valueIndex]
		if value != nil {
			columns, values := relationship.GetInsert(value)
			insertColumns = append(insertColumns, columns...)
			insertValues = append(insertValues, values...)
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

	w.WriteHeader(http.StatusCreated)
	// return created object as though it were a GET
	detailGET(w, r, c, m, newId, include)
}
