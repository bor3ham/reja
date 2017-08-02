package relationships

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
)

func relationLink(c schema.Context, modelType string, id string, key string) string {
	server := c.GetServer()
	request := c.GetRequest()
	url := fmt.Sprintf(
		"https://%s%s/%s/relationships/%s",
		request.Host,
		server.GetRoute(modelType),
		id,
		key,
	)
	return url
}
