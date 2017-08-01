package servers

import (
	"encoding/json"
	"github.com/bor3ham/reja/schema"
	"net/http"
)

func ParameterInfoHandler(
	s schema.Server,
	m *schema.Model,
	w http.ResponseWriter,
	r *http.Request,
) {
	rc := NewRequestContext(s, r)

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

	encoder := json.NewEncoder(w)
	if rc.GetServer().Whitespace() {
		encoder.SetIndent("", "    ")
	}
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(responseBlob)
	if err != nil {
		panic(err)
	}

	rc.LogStats()
}
