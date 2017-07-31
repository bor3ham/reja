package servers

import (
	"encoding/json"
	"fmt"
	"github.com/bor3ham/reja/schema"
	"github.com/davecgh/go-spew/spew"
	"io/ioutil"
	"net/http"
	"strings"
)

func detailPATCH(
	w http.ResponseWriter,
	r *http.Request,
	c schema.Context,
	m *schema.Model,
	id string,
	include *schema.Include,
) {
	// get instance
	noInclude := schema.Include{
		Children: map[string]*schema.Include{},
	}
	instances, _, err := c.GetObjectsByIDsAllRelations(m, []string{id}, &noInclude)
	if err != nil {
		panic(err)
	}
	if len(instances) == 0 {
		fmt.Fprintf(w, "No %s with that ID", m.Type)
		return
	}
	instance := instances[0]

	// read request data
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	// unmarshal request data into instance struct
	updatedInstance := m.Manager.Create()
	dataBlob := struct {
		Data interface{} `json:"data"`
	}{
		Data: updatedInstance,
	}
	err = json.Unmarshal(body, &dataBlob)
	if err != nil {
		BadRequest(c, w, "Unable to Parse JSON", err.Error())
		return
	}

	// extract existing values
	originalsMap := instance.GetValues()
	originals := valuesFromMap(originalsMap, m.Attributes, m.Relationships)
	// and updated values
	updatesMap := updatedInstance.GetValues()
	updates := valuesFromMap(updatesMap, m.Attributes, m.Relationships)
	// check all attributes for validity
	valueIndex := 0
	for _, attribute := range m.Attributes {
		if updates[valueIndex] != nil {
			updates[valueIndex], err = attribute.ValidateUpdate(updates[valueIndex], originals[valueIndex])
			if err != nil {
				BadRequest(c, w, "Bad Attribute Value", err.Error())
				return
			}
		}
		valueIndex += 1
	}
	// and all relationships
	for _, relation := range m.Relationships {
		if updates[valueIndex] != nil {
			updates[valueIndex], err = relation.ValidateUpdate(c, updates[valueIndex], originals[valueIndex])
			if err != nil {
				BadRequest(c, w, "Bad Relationship Value", err.Error())
				return
			}
		}
		valueIndex += 1
	}

	// build update query
	nextArgIndex := 1
	var updateKeys []string
	var updateArgs []interface{}

	valueIndex = 0
	for _, attribute := range m.Attributes {
		// skip nil values (use database default)
		value := updates[valueIndex]
		if value != nil {
			columns, values := attribute.GetInsert(value)
			for index, column := range columns {
				updateKeys = append(updateKeys, fmt.Sprintf("%s = $%d", column, nextArgIndex))
				updateArgs = append(updateArgs, values[index])
				nextArgIndex += 1
			}
		}
		valueIndex += 1
	}
	for _, relation := range m.Relationships {
		value := updates[valueIndex]
		if value != nil {
			columns, values := relation.GetInsert(value)
			for index, column := range columns {
				updateKeys = append(updateKeys, fmt.Sprintf("%s = $%d", column, nextArgIndex))
				updateArgs = append(updateArgs, values[index])
				nextArgIndex += 1
			}
		}
		valueIndex += 1
	}
	spew.Dump(updates)

	// start a transaction
	tx, err := c.Begin()
	if err != nil {
		panic(err)
	}

	updateQueries := []schema.Query{}
	valueIndex = len(m.Attributes)
	for _, relation := range m.Relationships {
		value := updates[valueIndex]
		if value != nil {
			original := originals[valueIndex]
			updateQueries = append(updateQueries, relation.GetUpdateQueries(id, original, value)...)
		}
		valueIndex += 1
	}

	for _, query := range updateQueries {
		tx.Exec(query.Query, query.Args...)
	}

	if len(updateKeys) > 0 {
		idArg := nextArgIndex
		nextArgIndex += 1
		query := fmt.Sprintf(
			`
				update %s
				set %s
				where %s = $%d
			`,
			m.Table,
			strings.Join(updateKeys, ", "),
			m.IDColumn,
			idArg,
		)

		_, err = tx.Exec(query, append(updateArgs, id)...)
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

	// flush instance cache
	c.FlushCache()
	// return updated object as though it were a GET
	detailGET(w, r, c, m, id, include)
}
