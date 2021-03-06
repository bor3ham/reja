package servers

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

func RelationHandler(
	s schema.Server,
	m *schema.Model,
	relationship schema.Relationship,
	w http.ResponseWriter,
	r *http.Request,
) {
	rc := NewRequestContext(s, w, r)
	defer catchExceptions(rc, w)()
	err := rc.Authenticate()
	if err != nil {
		return
	}

	// parse query strings
	queryStrings := r.URL.Query()
	// todo: condense into helper method (same as list_get.go)
	// get pagination stuff
	minPageSize := 1
	maxPageSize := rc.GetServer().GetMaximumDirectPageSize()
	pageSize, err := GetIntParam(
		queryStrings,
		"page[size]",
		"Page Size",
		rc.GetServer().GetDefaultDirectPageSize(),
		&minPageSize,
		&maxPageSize,
	)
	if err != nil {
		BadRequest(rc, w, "Bad Page Size Parameter", err.Error())
		return
	}
	minPageOffset := 1
	pageOffset, err := GetIntParam(
		queryStrings,
		"page[offset]",
		"Page Offset",
		1,
		&minPageOffset,
		nil,
	)
	if err != nil {
		BadRequest(rc, w, "Bad Page Offset Parameter", err.Error())
		return
	}
	offset := (pageOffset - 1) * pageSize

	// parse query strings
	_ = r.URL.Query()

	// extract id
	vars := mux.Vars(r)
	id := vars["id"]

	// get parent object
	instances, _, err := rc.GetObjectsByIDs(m, []string{id}, &schema.Include{})
	if err != nil {
		panic(err)
	}
	// abort if it doesn't exist
	if len(instances) == 0 {
		NotFound(rc, w, m.Type, id)
		return
	}

	extraColumns, _ := relationship.GetSelectExtra()
	var extraVariables [][]interface{}
	if len(extraColumns) > 0 {
		rows, err := rc.Query(fmt.Sprintf(
			`select %s from %s where %s = $1`,
			strings.Join(extraColumns, ", "),
			m.Table,
			m.IDColumn,
		), id)
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		for rows.Next() {
			_, vars := relationship.GetSelectExtra()
			rows.Scan(vars...)
			extraVariables = append(extraVariables, vars)
		}
	}
	values, _ := relationship.GetValues(rc, m, []string{id}, extraVariables, offset, pageSize)
	defaultValue := relationship.GetDefaultValue()
	var responseBlob interface{}
	responseBlob, exists := values[id]
	if !exists {
		responseBlob = defaultValue
	}

	rc.WriteToResponse(responseBlob)
	rc.LogStats()
}
