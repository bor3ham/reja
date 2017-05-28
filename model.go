package reja

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"github.com/bor3ham/reja/attributes"
	"github.com/bor3ham/reja/database"
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

type RelationResult struct{
	Values map[int]interface{}
	Default interface{}
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
	rows, err := database.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	instances := []Instance{}
	for rows.Next() {
		instance := m.Manager.Create()
		rows.Scan(instance.GetFields()...)
		instance.Clean()
		instances = append(instances, instance)
	}
	ids := []string{}

	for _, instance := range instances {
		ids = append(ids, strconv.Itoa(instance.GetID()))
	}
	keyed_values := []RelationResult{}
	for _, relationship := range m.Relationships {
		keyed_values = append(keyed_values, RelationResult{
			Values: relationship.GetKeyedValues(
				fmt.Sprintf("id in (%s)", strings.Join(ids, ", ")),
			),
			Default: relationship.GetEmptyKeyedValue(),
		})
	}
	for _, instance := range instances {
		instance_values := []interface{}{}
		for _, value := range keyed_values {
			item, exists := value.Values[instance.GetID()]
			if exists {
				instance_values = append(instance_values, item)
			} else {
				instance_values = append(instance_values, value.Default)
			}
		}

		instance.SetValues(instance_values)
	}

	general_instances := []interface{}{}
	for _, instance := range instances {
		general_instances = append(general_instances, instance)
	}
	response_data, err := json.Marshal(struct {
		Data []interface{} `json:"data"`
	}{
		Data: general_instances,
	})
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, string(response_data))
}

func (m Model) DetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	string_id := vars["id"]
	id, err := strconv.Atoi(string_id)
	if err != nil {
		panic(err)
	}
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
	err = database.QueryRow(query, id).Scan(instance.GetFields()...)

	keyed_values := []RelationResult{}
	for _, relationship := range m.Relationships {
		keyed_values = append(keyed_values, RelationResult{
			Values: relationship.GetKeyedValues("id = 1"),
			Default: relationship.GetEmptyKeyedValue(),
		})
	}
	instance_values := []interface{}{}
	for _, value := range keyed_values {
		item, exists := value.Values[id]
		if exists {
			instance_values = append(instance_values, item)
		} else {
			instance_values = append(instance_values, value.Default)
		}
	}
	instance.SetValues(instance_values)

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
