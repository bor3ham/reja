package models

import (
	"fmt"
	"github.com/bor3ham/reja/context"
	rejaHttp "github.com/bor3ham/reja/http"
	"github.com/gorilla/mux"
	"net/http"
)

func (m Model) DetailHandler(w http.ResponseWriter, r *http.Request) {
	rc := context.RequestContext{Request: r}
	queryStrings := r.URL.Query()

	// extract included information
	include, err := parseInclude(&m, queryStrings)
	if err != nil {
		rejaHttp.BadRequest(w, "Bad Included Relations Parameter", err.Error())
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	instances, included, err := GetObjects(&rc, m, []string{id}, 0, 0, include)
	if err != nil {
		panic(err)
	}

	if len(instances) == 0 {
		fmt.Fprintf(w, "No %s with that ID", m.Type)
		return
	}

	responseBlob := struct {
		Data interface{} `json:"data"`
		Included interface{} `json:"included,omitempty"`
	}{
		Data: instances[0],
	}
	if len(included) > 0 {
		var generalIncluded []interface{}
		for _, instance := range included {
			generalIncluded = append(generalIncluded, instance)
		}
		responseBlob.Included = generalIncluded
	}
	responseBytes := rejaHttp.MustJSONMarshal(responseBlob)
	fmt.Fprintf(w, string(responseBytes))
	logQueryCount(rc.GetQueryCount())
}
