package relationships

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
)

func relatedLink(c schema.Context, modelType string, id string, key string) string {
	server := c.GetServer()
	request := c.GetRequest()
	url := fmt.Sprintf(
		"%s%s/%s/%s",
		request.Host,
		server.GetRoute(modelType),
		id,
		key,
	)
	return url
}

func relationLink(c schema.Context, modelType string, id string, key string) string {
	server := c.GetServer()
	request := c.GetRequest()
	url := fmt.Sprintf(
		"%s%s/%s/relationships/%s",
		request.Host,
		server.GetRoute(modelType),
		id,
		key,
	)
	if server.GetIndirectPageSize() != server.GetDefaultDirectPageSize() {
		url += fmt.Sprintf("?page[size]=%d", server.GetIndirectPageSize())
	}
	return url
}
