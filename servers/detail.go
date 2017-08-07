package servers

import (
	"github.com/bor3ham/reja/schema"
	"github.com/gorilla/mux"
	"net/http"
)

func DetailHandler(s schema.Server, m *schema.Model, w http.ResponseWriter, r *http.Request) {
	rc := NewRequestContext(s, w, r)
	err := rc.Authenticate()
	if err != nil {
		return
	}

	// parse query strings
	queryStrings := r.URL.Query()

	// extract included information
	include, err := parseInclude(rc, m, queryStrings)
	if err != nil {
		BadRequest(rc, w, "Bad Included Relations Parameter", err.Error())
		return
	}

	// extract id
	vars := mux.Vars(r)
	id := vars["id"]

	// handle request based on method
	if r.Method == "PATCH" || r.Method == "PUT" {
		detailPATCH(w, r, rc, m, id, include)
	} else if r.Method == "GET" {
		detailGET(w, r, rc, m, id, include)
	} else {
		MethodNotAllowed(rc, w)
	}

	rc.LogStats()
}
