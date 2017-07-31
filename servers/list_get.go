package servers

import (
	"encoding/json"
	"fmt"
	"github.com/bor3ham/reja/schema"
	"github.com/bor3ham/reja/utils"
	"github.com/mailru/easyjson"
	"net/http"
	"strings"
	// "github.com/davecgh/go-spew/spew"
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
	validatedOrderParam := ""
	orderQueryArgs := []string{}
	validOrders := map[string]string{
		"id": m.IDColumn,
	}
	for _, attribute := range m.Attributes {
		attrOrders := attribute.GetOrderMap()
		for key, arg := range attrOrders {
			validOrders[key] = arg
		}
	}
	orders, err := GetStringParam(queryStrings, ORDER_ARG, "Ordering", m.DefaultOrder)
	if err != nil {
		BadRequest(c, w, "Bad Ordering Parameter", err.Error())
		return
	}
	splitOrders := strings.Split(orders, ",")
	orderedColumns := map[string]bool{}
	for _, order := range splitOrders {
		cleanOrder := strings.ToLower(strings.TrimSpace(order))
		if len(cleanOrder) == 0 {
			continue
		}
		posCleanOrder := strings.TrimPrefix(cleanOrder, "-")
		column, exists := validOrders[posCleanOrder]
		if !exists {
			BadRequest(c, w, "Bad Ordering Parameter", fmt.Sprintf(
				"Cannot order by '%s'.",
				cleanOrder,
			))
			return
		}
		_, exists = orderedColumns[column]
		if exists {
			BadRequest(c, w, "Bad Ordering Parameter", "Cannot order by the same column twice.")
			return
		}
		orderedColumns[column] = true
		query := column
		if posCleanOrder != cleanOrder {
			query += " desc"
		}
		orderQueryArgs = append(orderQueryArgs, query)
		if len(validatedOrderParam) != 0 {
			validatedOrderParam += ","
		}
		validatedOrderParam += cleanOrder
	}
	orderQuery := ""
	if len(orderQueryArgs) > 0 {
		orderQuery = fmt.Sprintf(
			"order by %s",
			strings.Join(orderQueryArgs, ", "),
		)
	}
	if validatedOrderParam == m.DefaultOrder {
		validatedOrderParam = ""
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

	if c.GetServer().UseEasyJSON() {
		_, _, err = easyjson.MarshalToHTTPResponseWriter(responseBlob, w)
	} else {
		encoder := json.NewEncoder(w)
		if c.GetServer().Whitespace() {
			encoder.SetIndent("", "    ")
		}
		encoder.SetEscapeHTML(false)
		err = encoder.Encode(responseBlob)
	}
	if err != nil {
		panic(err)
	}
}
