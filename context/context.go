package context

import (
	"database/sql"
	"github.com/bor3ham/reja/instances"
)

type Transaction struct {
	tx *sql.Tx
	c Context
}
func (t *Transaction) QueryRow(query string, args ...interface{}) *sql.Row {
	t.c.IncrementQueryCount()
	return t.tx.QueryRow(query, args...)
}
func (t *Transaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	t.c.IncrementQueryCount()
	return t.tx.Query(query, args...)
}
func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
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
	IncrementQueryCount()
	QueryRow(string, ...interface{}) *sql.Row
	Query(string, ...interface{}) (*sql.Rows, error)
	Exec(string, ...interface{}) (sql.Result, error)
	Begin() (*Transaction, error)

	InitCache()
	CacheObject(instances.Instance, map[string]map[string][]string)
	GetCachedObject(string, string) (instances.Instance, map[string]map[string][]string)
}
