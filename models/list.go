package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bor3ham/reja/database"
	rejaHttp "github.com/bor3ham/reja/http"
	"github.com/bor3ham/reja/instances"
	"math"
	"net/http"
	"strings"
)

const defaultPageSize = 5
const maximumPageSize = 400

func JSONMarshal(v interface{}, safeEncoding bool) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "    ")

	if safeEncoding {
		b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
		b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}
	return b, err
}

func badRequest(w http.ResponseWriter, title string, detail string) {
	errorBlob := struct {
		Exceptions []interface{} `json:"errors"`
	}{}
	errorBlob.Exceptions = append(errorBlob.Exceptions, struct {
		Title  string
		Detail string
	}{
		Title:  title,
		Detail: detail,
	})
	errorText, err := json.MarshalIndent(errorBlob, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, string(errorText))
}

func (m Model) ListHandler(w http.ResponseWriter, r *http.Request) {
	queryStrings := r.URL.Query()

	minPageSize := 1
	maxPageSize := maximumPageSize
	pageSize, err := rejaHttp.GetIntParam(
		queryStrings,
		"page[size]",
		"Page Size",
		defaultPageSize,
		&minPageSize,
		&maxPageSize,
	)
	if err != nil {
		badRequest(w, "Bad Page Size Parameter", err.Error())
		return
	}
	minPageOffset := 1
	pageOffset, err := rejaHttp.GetIntParam(
		queryStrings,
		"page[offset]",
		"Page Offset",
		1,
		&minPageOffset,
		nil,
	)
	if err != nil {
		badRequest(w, "Bad Page Offset Parameter", err.Error())
		return
	}
	offset := (pageOffset - 1) * pageSize

	countQuery := fmt.Sprintf(
		`
        select
            count(*)
        from %s
        `,
		m.Table,
	)
	var count int
	err = database.RequestQueryRow(r, countQuery).Scan(&count)
	if err != nil {
		panic(err)
	}
	lastPage := int(math.Ceil(float64(count) / float64(pageSize)))

	var nextUrl, prevUrl string
	if pageOffset < lastPage {
		nextUrl = r.Host + r.URL.Path
		nextUrl += fmt.Sprintf(`?page[size]=%d&page[offset]=%d`, pageSize, pageOffset+1)
	}
	if pageOffset > 1 {
		prevUrl = r.Host + r.URL.Path
		prevUrl += fmt.Sprintf(`?page[size]=%d&page[offset]=%d`, pageSize, pageOffset-1)
	}

	resultsQuery := fmt.Sprintf(
		`
      select
        %s,
        %s
      from %s
      limit %d
      offset %d
    `,
		m.IDColumn,
		strings.Join(m.FieldNames(), ","),
		m.Table,
		pageSize,
		offset,
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
			Values:  relationship.GetValues(r, ids),
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

	pageLinks := map[string]*string{}
	firstPageLink := r.Host + r.URL.Path
	pageLinks["first"] = &firstPageLink
	lastPageLink := r.Host + r.URL.Path + fmt.Sprintf(`?page[size]=%d&page[offset]=%d`, pageSize, lastPage)
	pageLinks["last"] = &lastPageLink
	pageLinks["prev"] = nil
	if prevUrl != "" {
		pageLinks["prev"] = &prevUrl
	}
	pageLinks["next"] = nil
	if nextUrl != "" {
		pageLinks["next"] = &nextUrl
	}
	pageMeta := map[string]int{}
	pageMeta["total"] = count
	pageMeta["count"] = len(instances)

	generalInstances := []interface{}{}
	for _, instance := range instances {
		generalInstances = append(generalInstances, instance)
	}

	responseBlob := struct {
		Links    interface{}   `json:"links"`
		Metadata interface{}   `json:"meta"`
		Data     []interface{} `json:"data"`
	}{
		Links:    pageLinks,
		Metadata: pageMeta,
		Data:     generalInstances,
	}

	response_data, err := JSONMarshal(responseBlob, true)
	if err != nil {
		panic(err)
	}

	logQueryCount(r)
	fmt.Fprintf(w, string(response_data))
}
