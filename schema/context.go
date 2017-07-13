package schema

import (
	"database/sql"
	"net/http"
)

type Context interface {
	GetServer() Server
	GetRequest() *http.Request

	IncrementQueryCount()
	QueryRow(string, ...interface{}) *sql.Row
	Query(string, ...interface{}) (*sql.Rows, error)
	Exec(string, ...interface{}) (sql.Result, error)
	Begin() (Transaction, error)

	InitCache()
	CacheObject(Instance, map[string]map[string][]string)
	GetCachedObject(string, string) (Instance, map[string]map[string][]string)
	GetObjectsByIDs(*Model, []string, *Include) ([]Instance, []Instance, error)
	GetObjectsByFilter(
		*Model,
		[]string,
		[]interface{},
		int,
		int,
		*Include,
	) (
		[]Instance,
		[]Instance,
		error,
	)
}
