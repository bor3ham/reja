package database

import (
	"database/sql"
	"fmt"
)

var config struct {
	Database *sql.DB
}

type QueryBlob struct {
	Query string
	Args  []interface{}
}

func InitialiseDatabase(database *sql.DB) {
	config.Database = database
}

func logQuery(query string) {
	fmt.Println(query)
}

func QueryRow(query string, args ...interface{}) *sql.Row {
	// logQuery(query)
	return config.Database.QueryRow(query, args...)
}

func Query(query string, args ...interface{}) (*sql.Rows, error) {
	// logQuery(query)
	return config.Database.Query(query, args...)
}

func Exec(query string, args ...interface{}) (sql.Result, error) {
	// logQuery(query)
	return config.Database.Exec(query, args...)
}

func Begin() (*sql.Tx, error) {
	return config.Database.Begin()
}
