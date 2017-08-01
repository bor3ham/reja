package servers

import (
	"github.com/bor3ham/reja/schema"
	"github.com/bor3ham/reja/utils"
	"net/http"
)

func ListHandler(s schema.Server, m *schema.Model, w http.ResponseWriter, r *http.Request) {
	rc := NewRequestContext(s, r)

	// get the authenticated user
	user, err := s.Authenticate(r)
	if err != nil {
		authError, ok := err.(utils.AuthError)
		if ok {
			w.WriteHeader(authError.Status)
		}
		return
	}
	rc.SetUser(user)

	// parse query strings
	queryStrings := r.URL.Query()

	// extract included information
	include, err := parseInclude(rc, m, queryStrings)
	if err != nil {
		BadRequest(rc, w, "Bad Included Relations Parameter", err.Error())
		return
	}

	// handle request based on method
	if r.Method == "POST" {
		listPOST(w, r, rc, m, queryStrings, include)
	} else if r.Method == "GET" {
		listGET(w, r, rc, m, queryStrings, include)
	}

	rc.LogStats()
}
