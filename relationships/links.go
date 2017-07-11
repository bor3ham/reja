package relationships

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
)

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
	return url
}
