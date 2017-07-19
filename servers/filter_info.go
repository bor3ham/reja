package servers

import (
	"encoding/json"
	"github.com/bor3ham/reja/schema"
	"net/http"
)

func FilterInfoHandler(
	s schema.Server,
	m *schema.Model,
	w http.ResponseWriter,
	r *http.Request,
) {
	// initialise request context
	rc := &RequestContext{
		Server:  s,
		Request: r,
	}
	rc.InitCache()

	filters := []string{}
	for _, attribute := range m.Attributes {
		filters = append(filters, attribute.AvailableFilters()...)
	}

	responseBlob := struct {
		Filters []string `json:"filters"`
	}{
		Filters: filters,
	}

	encoder := json.NewEncoder(w)
	if rc.GetServer().Whitespace() {
		encoder.SetIndent("", "    ")
	}
	err := encoder.Encode(responseBlob)
	if err != nil {
		panic(err)
	}
}
