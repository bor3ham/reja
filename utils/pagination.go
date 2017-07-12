package utils

import (
	"math"
	"strconv"
	"net/url"
)

// lazy encoded
const PAGE_SIZE = "page[size]"
const PAGE_OFFSET = "page[offset]"

func joinQueries(baseUrl string, queries ...map[string]string) string {
	fullUrl := baseUrl
	first := true
	for _, queryset := range queries {
		for key, query := range queryset {
			if first {
				fullUrl += "?"
				first = false
			} else {
				fullUrl += "&"
			}
			fullUrl += url.QueryEscape(key)
			fullUrl += "=" + url.QueryEscape(query)
		}
	}
	return fullUrl
}

func GetPaginationLinks(
	baseUrl string,
	currentPage int,
	pageSize int,
	defaultPageSize int,
	totalItems int,
	extraQueries map[string]string,
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
	selfLink := joinQueries(baseUrl, selfArgs, extraQueries)
	links["self"] = &selfLink

	// link to the first page
	firstArgs := map[string]string{}
	if pageSize != defaultPageSize {
		firstArgs[PAGE_SIZE] = strconv.Itoa(pageSize)
	}
	firstLink := joinQueries(baseUrl, firstArgs, extraQueries)
	links["first"] = &firstLink

	// link to the last page
	lastArgs := map[string]string{}
	if pageSize != defaultPageSize {
		lastArgs[PAGE_SIZE] = strconv.Itoa(pageSize)
	}
	if lastPage != 1 {
		lastArgs[PAGE_OFFSET] = strconv.Itoa(lastPage)
	}
	lastLink := joinQueries(baseUrl, lastArgs, extraQueries)
	links["last"] = &lastLink

	// link to the previous page
	links["prev"] = nil
	if currentPage > 1 {
		prevArgs := map[string]string{}
		if pageSize != defaultPageSize {
			prevArgs[PAGE_SIZE] = strconv.Itoa(pageSize)
		}
		prevArgs[PAGE_OFFSET] = strconv.Itoa(currentPage - 1)
		prevLink := joinQueries(baseUrl, prevArgs, extraQueries)
		links["prev"] = &prevLink
	}

	// link to the next page
	links["next"] = nil
	if currentPage < lastPage {
		nextArgs := map[string]string{}
		if pageSize != defaultPageSize {
			nextArgs[PAGE_SIZE] = strconv.Itoa(pageSize)
		}
		nextArgs[PAGE_OFFSET] = strconv.Itoa(currentPage + 1)
		nextLink := joinQueries(baseUrl, nextArgs, extraQueries)
		links["next"] = &nextLink
	}

	return links
}
