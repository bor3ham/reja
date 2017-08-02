package servers

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
	"github.com/bor3ham/reja/utils"
	"net/http"
	"strings"
)

const ORDER_ARG = "order"

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
		BadRequest(c, w, "Bad Page Size Parameter", err.Error())
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
		BadRequest(c, w, "Bad Page Offset Parameter", err.Error())
		return
	}
	offset := (pageOffset - 1) * pageSize

	// extract filters
	var validFilters []schema.Filter
	for _, attribute := range m.Attributes {
		filters, err := attribute.ValidateFilters(queryStrings)
		if err != nil {
			BadRequest(c, w, "Bad Filter Parameter", err.Error())
			return
		}
		validFilters = append(validFilters, filters...)
	}
	for _, relationship := range m.Relationships {
		filters, err := relationship.ValidateFilters(queryStrings)
		if err != nil {
			BadRequest(c, w, "Bad Filter Parameter", err.Error())
			return
		}
		validFilters = append(validFilters, filters...)
	}

	// create where clause from filters
	whereQueries := []string{}
	whereArgs := []interface{}{}
	for _, filter := range validFilters {
		queries, args := filter.GetWhere(c, m.Table, m.IDColumn, len(whereArgs)+1)

		whereQueries = append(whereQueries, queries...)
		whereArgs = append(whereArgs, args...)
	}
	whereClause := ""
	if len(whereQueries) > 0 {
		whereClause = fmt.Sprintf("where %s", strings.Join(whereQueries, " and "))
	}

	countQuery := fmt.Sprintf(
		`
			select
				count(*)
			from %s
			%s
		`,
		m.Table,
		whereClause,
	)
	var count int
	err = c.QueryRow(countQuery, whereArgs...).Scan(&count)
	if err != nil {
		panic(err)
	}

	// extract ordering
	orders, err := GetStringParam(queryStrings, ORDER_ARG, "Ordering", m.DefaultOrder)
	if err != nil {
		BadRequest(c, w, "Bad Ordering Parameter", err.Error())
		return
	}
	orderQuery, validatedOrderParam, err := m.GetOrderQuery(orders)
	if err != nil {
		BadRequest(c, w, "Bad Ordering Parameter", err.Error())
		return
	}

	instances, included, err := c.GetObjectsByFilter(
		m,
		whereQueries,
		whereArgs,
		orderQuery,
		offset,
		pageSize,
		include,
	)
	if err != nil {
		panic(err)
	}

	validQueries := map[string][]string{}
	if len(validatedOrderParam) > 0 {
		validQueries[ORDER_ARG] = []string{validatedOrderParam}
	}
	validIncludeQuery := include.AsString()
	if len(validIncludeQuery) > 0 {
		validQueries["include"] = []string{validIncludeQuery}
	}
	for _, filter := range validFilters {
		key := filter.GetQArgKey()
		values := filter.GetQArgValues()
		validQueries[key] = values
	}

	pageLinks := utils.GetPaginationLinks(
		"https://"+r.Host+r.URL.Path,
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

	c.WriteToResponse(responseBlob)
}
