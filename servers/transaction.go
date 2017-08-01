package servers

import (
	"database/sql"
)

type ContextTransaction struct {
	tx *sql.Tx
	rc *RequestContext
}

func (t *ContextTransaction) QueryRow(query string, args ...interface{}) *sql.Row {
	if t.rc.GetServer().LogSQL() {
		LogQuery(query)
	}
	t.rc.IncrementQueryCount()
	return t.tx.QueryRow(query, args...)
}
func (t *ContextTransaction) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if t.rc.GetServer().LogSQL() {
		LogQuery(query)
	}
	t.rc.IncrementQueryCount()
	return t.tx.Query(query, args...)
}
func (t *ContextTransaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	if t.rc.GetServer().LogSQL() {
		LogQuery(query)
	}
	t.rc.IncrementQueryCount()
	return t.tx.Exec(query, args...)
}
func (t *ContextTransaction) Commit() error {
	return t.tx.Commit()
}
func (t *ContextTransaction) Rollback() error {
	return t.tx.Rollback()
}
