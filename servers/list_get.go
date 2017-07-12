package servers

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
	"github.com/bor3ham/reja/utils"
	"net/http"
)

func listGET(
	w http.ResponseWriter,
	r *http.Request,
	c schema.Context,
	m *schema.Model,
	queryStrings map[string][]string,
	include *schema.Include,
) {
	minPageSize := 1
	maxPageSize := c.GetServer().GetMaximumDirectPageSize()
	pageSize, err := GetIntParam(
		queryStrings,
		"page[size]",
		"Page Size",
		c.GetServer().GetDefaultDirectPageSize(),
		&minPageSize,
		&maxPageSize,
	)
	if err != nil {
		BadRequest(w, "Bad Page Size Parameter", err.Error())
		return
	}
	minPageOffset := 1
	pageOffset, err := GetIntParam(
		queryStrings,
		"page[offset]",
		"Page Offset",
		1,
		&minPageOffset,
		nil,
	)
	if err != nil {
		BadRequest(w, "Bad Page Offset Parameter", err.Error())
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
	err = c.QueryRow(countQuery).Scan(&count)
	if err != nil {
		panic(err)
	}

	instances, included, err := c.GetObjects(m, []string{}, offset, pageSize, include)
	if err != nil {
		panic(err)
	}

	validQueries := map[string]string{}
	validIncludeQuery := include.AsString()
	if len(validIncludeQuery) > 0 {
		validQueries["include"] = validIncludeQuery
	}

	pageLinks := utils.GetPaginationLinks(
		r.Host+r.URL.Path,
		pageOffset,
		pageSize,
		c.GetServer().GetDefaultDirectPageSize(),
		count,
		validQueries,
	)

	pageMeta := map[string]interface{}{}
	pageMeta["total"] = count
	pageMeta["count"] = len(instances)

	generalInstances := []interface{}{}
	for _, instance := range instances {
		generalInstances = append(generalInstances, instance)
	}

	responseBlob := schema.Page{
		Links:    pageLinks,
		Metadata: pageMeta,
		Data:     generalInstances,
	}
	if len(included) > 0 {
		uniqueIncluded := UniqueInstances(included)
		var generalIncluded []interface{}
		for _, instance := range uniqueIncluded {
			generalIncluded = append(generalIncluded, instance)
		}
		responseBlob.Included = &generalIncluded
	}

	responseBytes := MustJSONMarshal(responseBlob)
	fmt.Fprint(w, string(responseBytes))
}
