package servers

import (
	"fmt"
	"github.com/bor3ham/reja/schema"
	"net/http"
	"log"
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

func NotFound(c schema.Context, w http.ResponseWriter, model string, id string) {
	errorBlob := Error{
		Exceptions: []Exception{
			Exception{
				Title:  "Not Found",
				Detail: fmt.Sprintf(
					"No %s found with ID '%s'.",
					model,
					id,
				),
			},
		},
	}
	w.WriteHeader(http.StatusNotFound)
	c.WriteToResponse(errorBlob)
}

func MethodNotAllowed(c schema.Context, w http.ResponseWriter) {
	errorBlob := Error{
		Exceptions: []Exception{
			Exception{
				Title: "Method Not Allowed",
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

func InternalServerError(c schema.Context, w http.ResponseWriter) {
	errorBlob := Error{
		Exceptions: []Exception{
			Exception{
				Title: "Internal Server Error",
				Detail: "Something went wrong. Please try again later.",
			},
		},
	}
	w.WriteHeader(http.StatusInternalServerError)
	c.WriteToResponse(errorBlob)
}

func catchExceptions(c schema.Context, w http.ResponseWriter) func() {
	return func() {
		if err := recover(); err != nil {
			InternalServerError(c, w)
			log.Printf("Runtime panic: %v", err)
		}
	}
}
