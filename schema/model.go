package schema

import (
	"net/http"
)

type Model interface {
	GetManager() Manager
	GetType() string
	GetIDColumn() string
	GetTable() string
	GetRelationships() []Relationship

	FieldColumns() []string
	FieldVariables() []interface{}
	ExtraColumns() []string
	ExtraVariables() [][]interface{}

	ListHandler(Server, http.ResponseWriter, *http.Request)
	DetailHandler(Server, http.ResponseWriter, *http.Request)
}
