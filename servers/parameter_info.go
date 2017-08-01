package servers

import (
	"github.com/bor3ham/reja/schema"
	"net/http"
)

func ParameterInfoHandler(
	s schema.Server,
	m *schema.Model,
	w http.ResponseWriter,
	r *http.Request,
) {
	rc := NewRequestContext(s, w, r)
	err := rc.Authenticate()
	if err != nil {
		return
	}

	filters := []interface{}{}
	for _, attribute := range m.Attributes {
		filters = append(filters, attribute.AvailableFilters()...)
	}
	for _, relationship := range m.Relationships {
		filters = append(filters, relationship.AvailableFilters()...)
	}

	responseBlob := struct {
		Filters []interface{} `json:"filters"`
	}{
		Filters: filters,
	}

	rc.WriteToResponse(responseBlob)
	rc.LogStats()
}
