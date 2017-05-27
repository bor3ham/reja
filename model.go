package reja

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/bor3ham/reja/attributes"
	"github.com/bor3ham/reja/relationships"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
)

type Model struct {
	Type          string
	Table         string
	IDColumn      string
	Attributes    []attributes.Attribute
	Relationships []relationships.Relationship
	Manager       Manager
}

func (m Model) FieldColumns() []string {
	var columns []string
	for _, attribute := range m.Attributes {
		columns = append(columns, attribute.GetColumns()...)
	}
	for _, relationship := range m.Relationships {
		columns = append(columns, relationship.GetColumns()...)
	}
	return columns
}

func (m Model) ListHandler(w http.ResponseWriter, r *http.Request) {
	query := fmt.Sprintf(
		`
      select
        %s,
        %s
      from %s
    `,
		m.IDColumn,
		strings.Join(m.FieldColumns(), ","),
		m.Table,
	)
	rows, err := Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	instances := make([]interface{}, 0)
	for rows.Next() {
		instance := m.Manager.Create()
		rows.Scan(instance.GetFields()...)
		instance.Clean()
		instances = append(instances, instance)
	}
	response_data, err := json.Marshal(struct {
		Data []interface{} `json:"data"`
	}{
		Data: instances,
	})
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, string(response_data))
}

func (m Model) DetailHandler(w http.ResponseWriter, r *http.Request) {
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
		strings.Join(m.FieldColumns(), ","),
		m.Table,
		m.IDColumn,
	)
	instance := m.Manager.Create()
	err := QueryRow(query, id).Scan(instance.GetFields()...)
	instance.Clean()

	switch {
	case err == sql.ErrNoRows:
		fmt.Fprintf(w, "No %s with that ID", m.Type)
	case err != nil:
		log.Fatal(err)
	default:
		response_data, err := json.Marshal(struct {
			Data interface{} `json:"data"`
		}{
			Data: instance,
		})
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(w, string(response_data))
	}
}
