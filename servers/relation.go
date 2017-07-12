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
	// initialise request context
	rc := &RequestContext{
		Server:  s,
		Request: r,
	}
	rc.InitCache()

	// parse query strings
	_ = r.URL.Query()

	// extract id
	vars := mux.Vars(r)
	id := vars["id"]

	// get parent object
	instances, _, err := rc.GetObjects(m, []string{id}, 0, 0, &schema.Include{})
	if err != nil {
		panic(err)
	}
	// abort if it doesn't exist
	if len(instances) == 0 {
		fmt.Fprintf(w, "No %s with that ID", m.Type)
		return
	}

	extraColumns := relationship.GetSelectExtraColumns()
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
			vars := relationship.GetSelectExtraVariables()
			rows.Scan(vars...)
			extraVariables = append(extraVariables, vars)
		}
	}
	values, _ := relationship.GetValues(rc, m, []string{id}, extraVariables)
	defaultValue := relationship.GetDefaultValue()
	var responseBlob interface{}
	responseBlob, exists := values[id]
	if !exists {
		responseBlob = defaultValue
	}
	responseBytes := MustJSONMarshal(responseBlob)
	fmt.Fprint(w, string(responseBytes))
}
