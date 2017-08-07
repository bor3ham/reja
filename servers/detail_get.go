package servers

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
	"net/http"
)

func detailGET(
	w http.ResponseWriter,
	r *http.Request,
	c schema.Context,
	m *schema.Model,
	id string,
	include *schema.Include,
) {
	instances, included, err := c.GetObjectsByIDs(m, []string{id}, include)
	if err != nil {
		panic(err)
	}

	if len(instances) == 0 {
		fmt.Fprintf(w, "No %s with that ID", m.Type)
		return
	}

	hasAccess := CanAccessAllInstances(c, instances)
	if !hasAccess {
		Forbidden(c, w, "Forbidden", "You do not have access to this object.")
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

	c.WriteToResponse(responseBlob)
}
