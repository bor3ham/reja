package models

import (
	// "database/sql"
	"fmt"
	"github.com/bor3ham/reja/context"
	rejaHttp "github.com/bor3ham/reja/http"
	"github.com/gorilla/mux"
	// "log"
	"net/http"
	// "strings"
)

func (m Model) DetailHandler(w http.ResponseWriter, r *http.Request) {
	rc := context.RequestContext{Request: r}

	vars := mux.Vars(r)
	id := vars["id"]

	instances, err := GetObjects(&rc, m, []string{id}, 0, 0, nil)
	if err != nil {
		panic(err)
	}

	if len(instances) == 0 {
		fmt.Fprintf(w, "No %s with that ID", m.Type)
		return
	}

	responseBytes := rejaHttp.MustJSONMarshal(struct {
		Data interface{} `json:"data"`
	}{
		Data: instances[0],
	})
	fmt.Fprintf(w, string(responseBytes))
	logQueryCount(rc.GetQueryCount())
}
