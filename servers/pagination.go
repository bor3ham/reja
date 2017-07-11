package servers

import (
	"math"
	"strconv"
	"net/url"
)

// // lazy encoded
// const PAGE_SIZE = "page[size]"
// const PAGE_OFFSET = "page[offset]"

// RFC 3986 compliant
const PAGE_SIZE = "page%%5Bsize%%5D"
const PAGE_OFFSET = "page%%5Boffset%%5D"

func joinQueries(baseUrl string, queries map[string]string) string {
	fullUrl := baseUrl
	first := true
	for key, query := range queries {
		if first {
			fullUrl += "?"
			first = false
		} else {
			fullUrl += "&"
		}
		fullUrl += key
		fullUrl += "=" + url.QueryEscape(query)
	}
	return fullUrl
}

func getPaginationLinks(
	baseUrl string,
	currentPage int,
	pageSize int,
	defaultPageSize int,
	totalItems int,
) (
	map[string]*string,
) {
	links := map[string]*string{}

	// calculate what the last page would be
	lastPage := int(math.Ceil(float64(totalItems) / float64(pageSize)))

	// link to this page
	selfArgs := map[string]string{}
	if pageSize != defaultPageSize {
		selfArgs[PAGE_SIZE] = strconv.Itoa(pageSize)
	}
	if currentPage != 1 {
		selfArgs[PAGE_OFFSET] = strconv.Itoa(currentPage)
	}
	selfLink := joinQueries(baseUrl, selfArgs)
	links["self"] = &selfLink

	// link to the first page
	firstArgs := map[string]string{}
	if pageSize != defaultPageSize {
		firstArgs[PAGE_SIZE] = strconv.Itoa(pageSize)
	}
	firstLink := joinQueries(baseUrl, firstArgs)
	links["first"] = &firstLink

	// link to the last page
	lastArgs := map[string]string{}
	if pageSize != defaultPageSize {
		lastArgs[PAGE_SIZE] = strconv.Itoa(pageSize)
	}
	if lastPage != 1 {
		lastArgs[PAGE_OFFSET] = strconv.Itoa(lastPage)
	}
	lastLink := joinQueries(baseUrl, lastArgs)
	links["last"] = &lastLink

	// link to the previous page
	if currentPage > 1 {
		prevArgs := map[string]string{}
		if pageSize != defaultPageSize {
			prevArgs[PAGE_SIZE] = strconv.Itoa(pageSize)
		}
		prevArgs[PAGE_OFFSET] = strconv.Itoa(currentPage - 1)
		prevLink := joinQueries(baseUrl, prevArgs)
		links["prev"] = &prevLink
	}

	// link to the next page
	if currentPage < lastPage {
		nextArgs := map[string]string{}
		if pageSize != defaultPageSize {
			nextArgs[PAGE_SIZE] = strconv.Itoa(pageSize)
		}
		nextArgs[PAGE_OFFSET] = strconv.Itoa(currentPage + 1)
		nextLink := joinQueries(baseUrl, nextArgs)
		links["next"] = &nextLink
	}

	return links
}
