package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/bor3ham/reja/context"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
)

func (m Model) DetailHandler(w http.ResponseWriter, r *http.Request) {
	rc := context.RequestContext{Request: r}

	vars := mux.Vars(r)
	id := vars["id"]
	query := fmt.Sprintf(
		`
      select
        %s,
        %s
      from %s
      where %s = $1
      limit 1
    `,
		m.IDColumn,
		strings.Join(m.FieldNames(), ","),
		m.Table,
		m.IDColumn,
	)

	fields := m.FieldVariables()
	scan_fields := []interface{}{}
	scan_fields = append(scan_fields, &id)
	scan_fields = append(scan_fields, fields...)
	err := rc.QueryRow(query, id).Scan(scan_fields...)

	switch {
	case err == sql.ErrNoRows:
		fmt.Fprintf(w, "No %s with that ID", m.Type)
	case err != nil:
		log.Fatal(err)
	default:
		instance := m.Manager.Create()
		instance.SetID(id)

		relation_values := []RelationResult{}
		for _, relationship := range m.Relationships {
			relation_values = append(relation_values, RelationResult{
				Values:  relationship.GetValues(&rc, []string{id}),
				Default: relationship.GetDefaultValue(),
			})
		}
		for _, value := range relation_values {
			item, exists := value.Values[id]
			if exists {
				fields = append(fields, item)
			} else {
				fields = append(fields, value.Default)
			}
		}
		instance.SetValues(fields)

		response_data, err := json.MarshalIndent(struct {
			Data interface{} `json:"data"`
		}{
			Data: instance,
		}, "", "    ")
		if err != nil {
			panic(err)
		}

		logQueryCount(rc.GetQueryCount())
		fmt.Fprintf(w, string(response_data))
	}
}
