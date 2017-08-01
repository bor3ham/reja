package servers

import (
	"encoding/json"
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
	instances, _, err := rc.GetObjectsByIDs(m, []string{id}, &schema.Include{})
	if err != nil {
		panic(err)
	}
	// abort if it doesn't exist
	if len(instances) == 0 {
		fmt.Fprintf(w, "No %s with that ID", m.Type)
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
	values, _ := relationship.GetValues(rc, m, []string{id}, extraVariables, false)
	defaultValue := relationship.GetDefaultValue()
	var responseBlob interface{}
	responseBlob, exists := values[id]
	if !exists {
		responseBlob = defaultValue
	}

	encoder := json.NewEncoder(w)
	if rc.GetServer().Whitespace() {
		encoder.SetIndent("", "    ")
	}
	err = encoder.Encode(responseBlob)
	if err != nil {
		panic(err)
	}
}
