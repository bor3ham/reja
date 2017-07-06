package server

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
)

func (m Model) ListHandler(s schema.Server, w http.ResponseWriter, r *http.Request) {
	// initialise request context
	rc := &RequestContext{
		Server:  s,
		Request: r,
	}
	rc.InitCache()

	// parse query strings
	queryStrings := r.URL.Query()

	// extract included information
	include, err := parseInclude(rc, &m, queryStrings)
	if err != nil {
		BadRequest(w, "Bad Included Relations Parameter", err.Error())
		return
	}

	// handle request based on method
	if r.Method == "POST" {
		listPOST(w, r, rc, m, queryStrings, include)
	} else if r.Method == "GET" {
		listGET(w, r, rc, m, queryStrings, include)
	}

	logQueryCount(rc.GetQueryCount())
}

func listPOST(
	w http.ResponseWriter,
	r *http.Request,
	c schema.Context,
	m Model,
	queryStrings map[string][]string,
	include *schema.Include,
) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	// parse the user input into instance data struct
	instance := m.Manager.Create()
	dataBlob := struct {
		Data interface{} `json:"data"`
	}{
		Data: instance,
	}
	err = JSONUnmarshal(body, &dataBlob)
	if err != nil {
		BadRequest(w, "Unable to Parse JSON", err.Error())
		return
	}

	// user cannot choose their own id
	if len(instance.GetID()) != 0 {
		BadRequest(w, "Bad Object Value", "ID's are assigned not chosen.")
		return
	}
	// type cannot be messed with
	instanceType := instance.GetType()
	if !(len(instanceType) == 0 || instanceType == m.Type) {
		BadRequest(w, "Bad Object Value", "Type does not match endpoint model.")
		return
	}

	// load defaults and validate values
	values := instance.GetValues()
	valueIndex := 0
	for _, attribute := range m.Attributes {
		values[valueIndex] = attribute.DefaultFallback(values[valueIndex], instance)
		// nil values are not included in the insert statement (use db default)
		if values[valueIndex] != nil {
			values[valueIndex], err = attribute.Validate(values[valueIndex])
			if err != nil {
				BadRequest(w, "Bad Attribute Value", err.Error())
				return
			}
		}
		valueIndex += 1
	}
	for _, relation := range m.Relationships {
		values[valueIndex] = relation.DefaultFallback(c, values[valueIndex], instance)
		// nil values are ignored
		if values[valueIndex] != nil {
			values[valueIndex], err = relation.Validate(c, values[valueIndex])
			if err != nil {
				BadRequest(w, "Bad Relationship Value", err.Error())
				return
			}
		}
		valueIndex += 1
	}

	// build insert query
	var insertColumns []string
	var insertValues []interface{}

	valueIndex = 0
	for _, attribute := range m.Attributes {
		// skip nil values (use database default)
		if values[valueIndex] != nil {
			insertColumns = append(insertColumns, attribute.GetInsertColumns(values[valueIndex])...)
			insertValues = append(insertValues, attribute.GetInsertValues(values[valueIndex])...)
		}
		valueIndex += 1
	}

	var valuePlaces []string
	for index, _ := range insertValues {
		valuePlaces = append(valuePlaces, fmt.Sprintf("$%d", index+1))
	}
	query := fmt.Sprintf(
		`insert into %s (%s) values (%s) returning %s;`,
		m.Table,
		strings.Join(insertColumns, ", "),
		strings.Join(valuePlaces, ", "),
		m.IDColumn,
	)

	// start a transaction
	tx, err := c.Begin()
	if err != nil {
		panic(err)
	}

	// execute insert query
	var newId string
	err = tx.QueryRow(query, insertValues...).Scan(&newId)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	// build additional queries
	var queries []schema.Query
	valueIndex = 0
	valueIndex += len(m.Attributes)
	for _, relationship := range m.Relationships {
		if values[valueIndex] != nil {
			queries = append(queries, relationship.GetInsertQueries(newId, values[valueIndex])...)
		}
		valueIndex += 1
	}

	// execute additional queries
	for _, query := range queries {
		_, err := tx.Exec(query.Query, query.Args...)
		if err != nil {
			tx.Rollback()
			panic(err)
		}
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		panic(err)
	}

	// return created object as though it were a GET
	detailGET(w, r, c, m, newId, include)
}

func listGET(
	w http.ResponseWriter,
	r *http.Request,
	c schema.Context,
	m Model,
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
	lastPage := int(math.Ceil(float64(count) / float64(pageSize)))

	var nextUrl, prevUrl string
	if pageOffset < lastPage {
		nextUrl = r.Host + r.URL.Path
		nextUrl += fmt.Sprintf(`?page[offset]=%d`, pageOffset+1)
		if pageSize != c.GetServer().GetDefaultDirectPageSize() {
			nextUrl += fmt.Sprintf("&page[size]=%d", pageSize)
		}
	}
	if pageOffset > 1 {
		prevUrl = r.Host + r.URL.Path
		prevUrl += fmt.Sprintf(`?page[offset]=%d`, pageOffset-1)
		if pageSize != c.GetServer().GetDefaultDirectPageSize() {
			prevUrl += fmt.Sprintf("&page[size]=%d", pageSize)
		}
	}

	instances, included, err := c.GetObjects(&m, []string{}, offset, pageSize, include)
	if err != nil {
		panic(err)
	}

	pageLinks := map[string]*string{}
	firstPageLink := r.Host + r.URL.Path
	pageLinks["first"] = &firstPageLink
	lastPageLink := r.Host + r.URL.Path
	lastPageLink += fmt.Sprintf(`?page[offset]=%d`, lastPage)
	if pageSize != c.GetServer().GetDefaultDirectPageSize() {
		lastPageLink += fmt.Sprintf(`&page[size]=%d`, pageSize)
	}
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
	fmt.Fprintf(w, string(responseBytes))
}
