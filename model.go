package reja

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Model struct {
	Type       string
	Table      string
	IDColumn   string
	Attributes []Field
	Manager    Manager
}

func (m Model) FieldColumns() []string {
	var columns []string
	for _, attribute := range m.Attributes {
		columns = append(columns, attribute.ColumnName)
	}
	return columns
}

func (m Model) ListHandler(w http.ResponseWriter, r *http.Request) {
	instances := make([]interface{}, 0)
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
	query := fmt.Sprintf(
		`
      select
        %s,
        %s
      from %s
      where %s = $1
    `,
		m.IDColumn,
		strings.Join(m.FieldColumns(), ","),
		m.Table,
		m.IDColumn,
	)
	fmt.Println(query)
	instance := m.Manager.Create()
	err := Database.QueryRow(query, 1).Scan(instance.GetFields()...)

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
