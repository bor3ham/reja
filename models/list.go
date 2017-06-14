package models

import (
    "encoding/json"
    "fmt"
    "github.com/bor3ham/reja/database"
    "github.com/bor3ham/reja/instances"
    "net/http"
    "strings"
    "strconv"
)

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
    instances := []instances.Instance{}
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
