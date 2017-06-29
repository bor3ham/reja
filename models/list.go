package models

import (
	"fmt"
	"github.com/bor3ham/reja/context"
	"github.com/bor3ham/reja/format"
	rejaHttp "github.com/bor3ham/reja/http"
	"github.com/davecgh/go-spew/spew"
	"io/ioutil"
	"math"
	"net/http"
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
	rc.InitCache()

	if r.Method == "POST" {
		postList(w, r, &rc, m)
	} else if r.Method == "GET" {
		getList(w, r, &rc, m)
	}

	logQueryCount(rc.GetQueryCount())
}

func postList(w http.ResponseWriter, r *http.Request, rc context.Context, m Model) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	// parse the user input into instance data struct
	instance := m.Manager.Create()
	dataBlob := struct{
		Data interface{} `json:"data"`
	}{
		Data: instance,
	}
	err = rejaHttp.JSONUnmarshal(body, &dataBlob)
	if err != nil {
		rejaHttp.BadRequest(w, "Unable to Parse JSON", err.Error())
		return
	}

	// user cannot choose their own id
	if len(instance.GetID()) != 0 {
		rejaHttp.BadRequest(w, "Bad Object Value", "ID's are assigned not chosen.")
		return
	}
	// type cannot be messed with
	instanceType := instance.GetType()
	if !(len(instanceType) == 0 || instanceType == m.Type) {
		rejaHttp.BadRequest(w, "Bad Object Value", "Type does not match endpoint model.")
		return
	}

	// validate values
	values := instance.GetValues()
	valueIndex := 0
	for _, attribute := range m.Attributes {
		values[valueIndex], err = attribute.ValidateNew(values[valueIndex])
		if err != nil {
			rejaHttp.BadRequest(w, "Bad Attribute Value", err.Error())
			return
		}
		valueIndex += 1
	}
	for _, relation := range m.Relationships {
		values[valueIndex], err = relation.ValidateNew(values[valueIndex])
		if err != nil {
			rejaHttp.BadRequest(w, "Bad Relationship Value", err.Error())
			return
		}
		valueIndex += 1
	}
	// instance.SetValues(values)
	spew.Dump(instance)
	spew.Dump(values)

	fmt.Fprintf(w, "Post valid")
}

func getList(w http.ResponseWriter, r *http.Request, rc context.Context, m Model) {
	queryStrings := r.URL.Query()

	// extract included information
	include, err := parseInclude(&m, queryStrings)
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

	instances, included, err := GetObjects(rc, m, []string{}, offset, pageSize, include)
	if err != nil {
		panic(err)
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
	if len(included) > 0 {
		uniqueIncluded := UniqueInstances(included)
		var generalIncluded []interface{}
		for _, instance := range uniqueIncluded {
			generalIncluded = append(generalIncluded, instance)
		}
		responseBlob.Included = &generalIncluded
	}

	responseBytes := rejaHttp.MustJSONMarshal(responseBlob)
	fmt.Fprintf(w, string(responseBytes))
}
