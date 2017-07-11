package servers

import (
	"math"
	"github.com/google/go-querystring/query"
)

type PageArgs struct {
	Offset *int `url:"page[offset],omitempty"`
	Size *int `url:"page[size],omitempty"`
}

func mustJoinQueries(baseUrl string, args PageArgs) string {
	values, err := query.Values(args)
	if err != nil {
		panic(err)
	}
	encoded := values.Encode()
	if len(encoded) > 0 {
		return baseUrl + "?" + encoded
	}
	return baseUrl
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

	var linkPageSize *int
	if pageSize != defaultPageSize {
		linkPageSize = &pageSize
	}

	// link to this page
	selfArgs := PageArgs{
		Size: linkPageSize,
	}
	if currentPage != 1 {
		selfArgs.Offset = &currentPage
	}
	selfLink := mustJoinQueries(baseUrl, selfArgs)
	links["self"] = &selfLink

	// link to the first page
	firstArgs := PageArgs{
		Size: linkPageSize,
	}
	firstLink := mustJoinQueries(baseUrl, firstArgs)
	links["first"] = &firstLink

	// link to the last page
	lastArgs := PageArgs{
		Size: linkPageSize,
	}
	if lastPage != 1 {
		lastArgs.Offset = &lastPage
	}
	lastLink := mustJoinQueries(baseUrl, lastArgs)
	links["last"] = &lastLink

	// link to the next page
	if currentPage < lastPage {
		nextPage := currentPage + 1
		nextArgs := PageArgs{
			Size: linkPageSize,
			Offset: &nextPage,
		}
		nextLink := mustJoinQueries(baseUrl, nextArgs)
		links["next"] = &nextLink
	}

	// link to the previous page
	if currentPage > 0 {
		prevPage := currentPage - 1
		prevArgs := PageArgs{
			Size: linkPageSize,
			Offset: &prevPage,
		}
		prevLink := mustJoinQueries(baseUrl, prevArgs)
		links["prev"] = &prevLink
	}

	return links
}
