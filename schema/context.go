package schema

import (
	"database/sql"
)

type Context interface {
	GetServer() Server

	IncrementQueryCount()
	QueryRow(string, ...interface{}) *sql.Row
	Query(string, ...interface{}) (*sql.Rows, error)
	Exec(string, ...interface{}) (sql.Result, error)
	Begin() (Transaction, error)

	InitCache()
	CacheObject(Instance, map[string]map[string][]string)
	GetCachedObject(string, string) (Instance, map[string]map[string][]string)
	GetObjects(Model, []string, int, int, *Include) ([]Instance, []Instance, error)
}
