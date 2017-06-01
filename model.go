package reja

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	Values map[string]interface{}
	Default interface{}
}

func (m Model) FieldVariables() []interface{} {
	var fields []interface{}
	for _, attribute := range m.Attributes {
		fields = append(fields, attribute.GetColumnVariables()...)
	}
	for _, relationship := range m.Relationships {
		fields = append(fields, relationship.GetColumnVariables()...)
	}
	return fields
}

func (m Model) FieldNames() []string {
	var columns []string
	for _, attribute := range m.Attributes {
		columns = append(columns, attribute.GetColumnNames()...)
	}
	for _, relationship := range m.Relationships {
		columns = append(columns, relationship.GetColumnNames()...)
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
		strings.Join(m.FieldNames(), ","),
		m.Table,
	)
	rows, err := database.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	ids := []string{}
	instances := []Instance{}
	instance_fields := [][]interface{}{}
	for rows.Next() {
		var id string
		fields := m.FieldVariables()
		instance_fields = append(instance_fields, fields)
		scan_fields := []interface{}{}
		scan_fields = append(scan_fields, &id)
		scan_fields = append(scan_fields, fields...)
		err := rows.Scan(scan_fields...)
		if err != nil {
			panic(err)
		}

		instance := m.Manager.Create()
		instance.SetID(id)
		instances = append(instances, instance)

		ids = append(ids, id)
	}

	relation_values := []RelationResult{}
	for _, relationship := range m.Relationships {
		relation_values = append(relation_values, RelationResult{
			Values: relationship.GetValues(ids),
			Default: relationship.GetDefaultValue(),
		})
	}
	for instance_index, instance := range instances {
		for _, value := range relation_values {
			item, exists := value.Values[instance.GetID()]
			if exists {
				instance_fields[instance_index] = append(instance_fields[instance_index], item)
			} else {
				instance_fields[instance_index] = append(instance_fields[instance_index], value.Default)
			}
		}
	}

	for instance_index, instance := range instances {
		instance.SetValues(instance_fields[instance_index])
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
	err := database.QueryRow(query, id).Scan(scan_fields...)

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
				Values: relationship.GetValues([]string{id}),
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
