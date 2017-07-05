package context

import (
	"database/sql"
	"github.com/bor3ham/reja/database"
	"github.com/bor3ham/reja/instances"
	"github.com/bor3ham/reja/models"
)

type Transaction struct {
	tx *sql.Tx
	c  Context
}

func (t *Transaction) QueryRow(query string, args ...interface{}) *sql.Row {
	database.LogQuery(query)
	t.c.IncrementQueryCount()
	return t.tx.QueryRow(query, args...)
}
func (t *Transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	database.LogQuery(query)
	t.c.IncrementQueryCount()
	return t.tx.Query(query, args...)
}
func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	database.LogQuery(query)
	t.c.IncrementQueryCount()
	return t.tx.Exec(query, args...)
}
func (t *Transaction) Commit() error {
	return t.tx.Commit()
}
func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

type CachedInstance struct {
	Instance    instances.Instance
	RelationMap map[string]map[string][]string
}
type Context interface {
	GetServer() Server

	IncrementQueryCount()
	QueryRow(string, ...interface{}) *sql.Row
	Query(string, ...interface{}) (*sql.Rows, error)
	Exec(string, ...interface{}) (sql.Result, error)
	Begin() (*Transaction, error)

	InitCache()
	CacheObject(instances.Instance, map[string]map[string][]string)
	GetCachedObject(string, string) (instances.Instance, map[string]map[string][]string)
}

type Server interface {
	GetDatabase() *sql.DB

	GetDefaultDirectPageSize() int
	GetMaximumDirectPageSize() int
	GetIndirectPageSize() int

	GetModel(string) *models.Model
}
