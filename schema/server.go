package schema

import (
	"net/http"
)

type Server interface {
	GetDatabase() Database

	GetDefaultDirectPageSize() int
	GetMaximumDirectPageSize() int
	GetIndirectPageSize() int

	GetModel(string) *Model
	GetRoute(string) string

	Whitespace() bool
	UseEasyJSON() bool
	LogSQL() bool

	Authenticate(http.ResponseWriter, *http.Request) (User, error)
}
