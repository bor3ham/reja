package models

import (
	"database/sql"
	"github.com/bor3ham/reja/instances"
	"github.com/bor3ham/reja/database"
)

type Server interface {
	GetDatabase() *sql.DB

	GetDefaultDirectPageSize() int
	GetMaximumDirectPageSize() int
	GetIndirectPageSize() int

	GetModel(string) *Model
}

type Context interface {
	GetServer() Server

	IncrementQueryCount()
	QueryRow(string, ...interface{}) *sql.Row
	Query(string, ...interface{}) (*sql.Rows, error)
	Exec(string, ...interface{}) (sql.Result, error)
	Begin() (database.Transaction, error)

	InitCache()
	CacheObject(instances.Instance, map[string]map[string][]string)
	GetCachedObject(string, string) (instances.Instance, map[string]map[string][]string)
}
