package server

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
	"github.com/gorilla/mux"
	"net/http"
)

func (m Model) DetailHandler(s schema.Server, w http.ResponseWriter, r *http.Request) {
	// initialise request context
	rc := &RequestContext{
		Server:  s,
		Request: r,
	}
	rc.InitCache()

	// parse query strings
	queryStrings := r.URL.Query()

	// extract included information
	include, err := parseInclude(rc, &m, queryStrings)
	if err != nil {
		BadRequest(w, "Bad Included Relations Parameter", err.Error())
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
	}

	// log request stats
	logQueryCount(rc.GetQueryCount())
}

func detailPATCH(
	w http.ResponseWriter,
	r *http.Request,
	c schema.Context,
	m Model,
	id string,
	include *schema.Include,
) {

}

func detailGET(
	w http.ResponseWriter,
	r *http.Request,
	c schema.Context,
	m Model,
	id string,
	include *schema.Include,
) {
	instances, included, err := c.GetObjects(&m, []string{id}, 0, 0, include)
	if err != nil {
		panic(err)
	}

	if len(instances) == 0 {
		fmt.Fprintf(w, "No %s with that ID", m.Type)
		return
	}

	responseBlob := struct {
		Data     interface{} `json:"data"`
		Included interface{} `json:"included,omitempty"`
	}{
		Data: instances[0],
	}
	if len(included) > 0 {
		uniqueIncluded := UniqueInstances(included)
		var generalIncluded []interface{}
		for _, instance := range uniqueIncluded {
			generalIncluded = append(generalIncluded, instance)
		}
		responseBlob.Included = generalIncluded
	}
	responseBytes := MustJSONMarshal(responseBlob)
	fmt.Fprintf(w, string(responseBytes))
}
