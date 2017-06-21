package models

import (
	"fmt"
	"github.com/bor3ham/reja/context"
	"github.com/bor3ham/reja/format"
	rejaHttp "github.com/bor3ham/reja/http"
	"github.com/bor3ham/reja/instances"
	"math"
	"net/http"
	"strings"
	"github.com/davecgh/go-spew/spew"
)

const defaultPageSize = 5
const maximumPageSize = 400

func flattened(fields [][]interface{}) []interface{} {
	var flatList []interface{}
	for _, relation := range fields {
		flatList = append(flatList, relation...)
	}
	return flatList
}

func (m Model) ListHandler(w http.ResponseWriter, r *http.Request) {
	rc := context.RequestContext{Request: r}
	queryStrings := r.URL.Query()

	// extract included information
	_, err := parseInclude(&m, queryStrings)
	if err != nil {
		rejaHttp.BadRequest(w, "Bad Included Relations Parameter", err.Error())
		return
	}

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
		rejaHttp.BadRequest(w, "Bad Page Size Parameter", err.Error())
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
		rejaHttp.BadRequest(w, "Bad Page Offset Parameter", err.Error())
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
	err = rc.QueryRow(countQuery).Scan(&count)
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

	columns := m.FieldNames()
	columns = append(m.FieldNames(), m.ExtraNames()...)
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
		strings.Join(columns, ","),
		m.Table,
		pageSize,
		offset,
	)
	rows, err := rc.Query(resultsQuery)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	ids := []string{}
	instances := []instances.Instance{}
	instanceFields := [][]interface{}{}
	extraFields := [][][]interface{}{}
	for rows.Next() {
		var id string
		fields := m.FieldVariables()
		instanceFields = append(instanceFields, fields)
		extras := m.ExtraVariables()
		extraFields = append(extraFields, extras)
		flatExtras := flattened(extras)

		scanFields := []interface{}{}
		scanFields = append(scanFields, &id)
		scanFields = append(scanFields, fields...)
		scanFields = append(scanFields, flatExtras...)
		err := rows.Scan(scanFields...)
		if err != nil {
			panic(err)
		}

		instance := m.Manager.Create()
		instance.SetID(id)
		instances = append(instances, instance)

		ids = append(ids, id)
	}

	// relation map
	_ = map[string]map[string][]int{}

	relationValues := []RelationResult{}
	for relationIndex, relationship := range m.Relationships {
		values, relationIds := relationship.GetValues(&rc, ids, extraFields[relationIndex])
		spew.Dump(relationIds)
		relationValues = append(relationValues, RelationResult{
			Values:  values,
			Default: relationship.GetDefaultValue(),
		})
	}
	for instance_index, instance := range instances {
		for _, value := range relationValues {
			item, exists := value.Values[instance.GetID()]
			if exists {
				instanceFields[instance_index] = append(instanceFields[instance_index], item)
			} else {
				instanceFields[instance_index] = append(instanceFields[instance_index], value.Default)
			}
		}
	}

	for instance_index, instance := range instances {
		instance.SetValues(instanceFields[instance_index])
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
	pageMeta := map[string]interface{}{}
	pageMeta["total"] = count
	pageMeta["count"] = len(instances)

	generalInstances := []interface{}{}
	for _, instance := range instances {
		generalInstances = append(generalInstances, instance)
	}

	responseBlob := format.Page{
		Links:    pageLinks,
		Metadata: pageMeta,
		Data:     generalInstances,
	}

	responseBytes := rejaHttp.MustJSONMarshal(responseBlob)
	fmt.Fprintf(w, string(responseBytes))
	logQueryCount(rc.GetQueryCount())
}
