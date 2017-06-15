package context

import (
	"github.com/bor3ham/reja/database"
	"database/sql"
	gorillaContext "github.com/gorilla/context"
	"net/http"
)

type Context interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type RequestContext struct {
	Request *http.Request
}
func (rc *RequestContext) incrementQueryCount() {
	queries := rc.GetQueryCount()
	queries += 1
	gorillaContext.Set(rc.Request, "queries", queries)
}
func (rc *RequestContext) GetQueryCount() int {
	current := gorillaContext.Get(rc.Request, "queries")
	if current != nil {
		currentInt, ok := current.(int)
		if !ok {
			panic("Unable to convert query count to integer")
		}
		return currentInt
	}
	return 0
}
func (rc *RequestContext) QueryRow(query string, args ...interface{}) *sql.Row {
	rc.incrementQueryCount()
	return database.QueryRow(query, args...)
}
func (rc *RequestContext) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rc.incrementQueryCount()
	return database.Query(query, args...)
}
