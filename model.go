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
	"strconv"
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

func logQueryCount(r *http.Request) {
	num_queries := database.GetRequestQueryCount(r)
	fmt.Println("Database queries:", num_queries)
}

func (m Model) ListHandler(w http.ResponseWriter, r *http.Request) {
	queryStrings := r.URL.Query()

	var pageSize int
	pageSizeQueries, ok := queryStrings["page[size]"]
	if ok {
		if len(pageSizeQueries) == 0 {
			panic("Empty page size argument given")
		}
		if len(pageSizeQueries) > 1 {
			panic("Too many page size arguments given")
		}
		var err error
		pageSize, err = strconv.Atoi(pageSizeQueries[0])
		if err != nil {
			panic("Invalid page size argument given")
		}
		if pageSize < 1 {
			panic("Page size given less than 1")
		}
		if pageSize > 400 {
			panic("Page size given greater than 400")
		}
	} else {
		pageSize = 5
	}

	countQuery := fmt.Sprintf(
		`
		select
			count(*)
		from %s
		`,
		m.Table,
	)
	var count int
	err := database.RequestQueryRow(r, countQuery).Scan(&count)
	if err != nil {
		panic(err)
	}

	resultsQuery := fmt.Sprintf(
		`
      select
        %s,
        %s
      from %s
      limit %d
    `,
		m.IDColumn,
		strings.Join(m.FieldNames(), ","),
		m.Table,
		pageSize,
	)
	rows, err := database.RequestQuery(r, resultsQuery)
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
			Values: relationship.GetValues(r, ids),
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

	page_links := map[string]string{}
	current_url := "http://here"
	page_links["first"] = fmt.Sprintf("%s", current_url)
	// page_links["last"] = fmt.Sprintf("%s?last", current_url)
	// page_links["prev"] = fmt.Sprintf("%s?prev", current_url)
	// page_links["next"] = fmt.Sprintf("%s?next", current_url)
	page_meta := map[string]int{}
	page_meta["total"] = count
	page_meta["count"] = len(instances)

	general_instances := []interface{}{}
	for _, instance := range instances {
		general_instances = append(general_instances, instance)
	}
	response_data, err := json.MarshalIndent(struct {
		Links interface{} `json:"links"`
		Metadata interface{} `json:"meta"`
		Data []interface{} `json:"data"`
	}{
		Links: page_links,
		Metadata: page_meta,
		Data: general_instances,
	}, "", "    ")
	if err != nil {
		panic(err)
	}

	logQueryCount(r)
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
	err := database.RequestQueryRow(r, query, id).Scan(scan_fields...)

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
				Values: relationship.GetValues(r, []string{id}),
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

		logQueryCount(r)
		fmt.Fprintf(w, string(response_data))
	}
}
