package servers

import (
	"github.com/bor3ham/reja/schema"
	"net/http"
	"fmt"
)

type Error struct {
	Exceptions []Exception `json:"errors"`
}
type Exception struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func BadRequest(c schema.Context, w http.ResponseWriter, title string, detail string) {
	errorBlob := Error{
		Exceptions: []Exception{
			Exception{
				Title:  title,
				Detail: detail,
			},
		},
	}
	w.WriteHeader(http.StatusBadRequest)
	c.WriteToResponse(errorBlob)
}

func Forbidden(c schema.Context, w http.ResponseWriter, title string, detail string) {
	errorBlob := Error{
		Exceptions: []Exception{
			Exception{
				Title:  title,
				Detail: detail,
			},
		},
	}
	w.WriteHeader(http.StatusForbidden)
	c.WriteToResponse(errorBlob)
}

func MethodNotAllowed(c schema.Context, w http.ResponseWriter) {
	errorBlob := Error{
		Exceptions: []Exception{
			Exception{
				Title:  "Method Not Allowed",
				Detail: fmt.Sprintf(
					"This endpoint does not support %s requests.",
					c.GetRequest().Method,
				),
			},
		},
	}
	w.WriteHeader(http.StatusForbidden)
	c.WriteToResponse(errorBlob)
}
